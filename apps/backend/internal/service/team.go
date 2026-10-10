package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"net/http"
	"regexp"

	"github.com/6sLOGAN78/flux/internal/errs"
	"github.com/6sLOGAN78/flux/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ChangeRole validates canonical input and keeps authorization in the effect transaction.
func (s *TeamService) ChangeRole(ctx context.Context, scope repository.Scope, target uuid.UUID,
	role, key string,
) (repository.TeamMember, error) {
	keyOK, err := regexp.MatchString(`^[A-Za-z0-9_-]{16,128}$`, key)
	if err != nil || !keyOK || target == uuid.Nil || !Allows(role, CapabilityRead) {
		return repository.TeamMember{}, errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	if s == nil || s.workspace == nil {
		return repository.TeamMember{}, workspaceFailure(errors.New("team unavailable"))
	}
	hash := sha256.Sum256([]byte(target.String() + "\n" + role))
	ctx, cancel := context.WithTimeout(ctx, workspaceTimeout)
	defer cancel()
	result, err := s.store.ChangeRole(ctx, scope, target, role, key, hash[:],
		func(ctx context.Context, tx pgx.Tx, scope repository.Scope) (string, error) {
			if capabilityErr := s.workspace.RequireCapability(ctx, tx, scope, CapabilityTeam); capabilityErr != nil {
				return "", capabilityErr
			}
			var current string
			readErr := tx.QueryRow(ctx, "SELECT role FROM memberships WHERE workspace_id=$1 AND user_id=$2 FOR UPDATE",
				scope.WorkspaceID, scope.ActorID).Scan(&current)
			if readErr != nil {
				return "", workspaceFailure(readErr)
			}
			if !Allows(current, CapabilityTeam) {
				return "", errs.NewForbiddenError("You do not have permission for this action.", false)
			}
			return current, nil
		}, AllowsRoleChange)
	if errors.Is(err, repository.ErrOwnerRequired) {
		return repository.TeamMember{}, &errs.HTTPError{Code: "OWNER_REQUIRED",
			Message: "Promote another owner first", Status: http.StatusConflict}
	}
	if errors.Is(err, repository.ErrTeamForbidden) {
		return repository.TeamMember{}, errs.NewForbiddenError("You do not have permission for this action.", false)
	}
	var denied *errs.HTTPError
	if errors.As(err, &denied) {
		return repository.TeamMember{}, denied
	}
	return result, workspaceFailure(err)
}

// TeamService requires current SQL Team capability on every page.
type TeamService struct {
	store     *repository.TeamRepository
	workspace *WorkspaceService
}

// NewTeamService injects the scoped store and shared capability authority.
func NewTeamService(store *repository.TeamRepository, workspace *WorkspaceService) *TeamService {
	return &TeamService{store: store, workspace: workspace}
}

// List exposes at most 25 identities within the existing 64 KiB response budget.
func (s *TeamService) List(ctx context.Context, scope repository.Scope,
	after uuid.UUID,
) ([]repository.TeamMember, *string, error) {
	if s == nil || s.workspace == nil {
		return nil, nil, workspaceFailure(errors.New("team unavailable"))
	}
	ctx, cancel := context.WithTimeout(ctx, workspaceTimeout)
	defer cancel()
	items, err := s.store.List(ctx, scope, after, func(ctx context.Context, tx pgx.Tx, scope repository.Scope) error {
		return s.workspace.RequireCapability(ctx, tx, scope, CapabilityTeam)
	})
	if err != nil {
		var denied *errs.HTTPError
		if errors.As(err, &denied) {
			return nil, nil, denied
		}
		return nil, nil, workspaceFailure(err)
	}
	var next *string
	const pageSize = 25
	if len(items) > pageSize {
		items = items[:pageSize]
		position := items[pageSize-1].ID.String()
		next = &position
	}
	return items, next, nil
}
