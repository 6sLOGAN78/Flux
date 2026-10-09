package service

import (
	"context"
	"errors"

	"github.com/6sLOGAN78/flux/internal/errs"
	"github.com/6sLOGAN78/flux/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

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
