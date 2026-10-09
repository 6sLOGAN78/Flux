package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/6sLOGAN78/flux/internal/errs"
	"github.com/6sLOGAN78/flux/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// WorkspaceService validates bootstrap requests and enforces Flux membership.
type WorkspaceService struct {
	store *repository.WorkspaceRepository
}

// NewWorkspaceService injects the authoritative tenant store.
func NewWorkspaceService(store *repository.WorkspaceRepository) *WorkspaceService {
	return &WorkspaceService{store: store}
}

const workspaceTimeout = 3 * time.Second

// Create requires an explicit Unicode name and identity-scoped durable retry key.
func (s *WorkspaceService) Create(ctx context.Context, actor uuid.UUID,
	name, key string,
) (repository.Workspace, error) {
	name = strings.TrimSpace(name)
	keyOK, err := regexp.MatchString(`^[A-Za-z0-9_-]{16,128}$`, key)
	if err != nil || !keyOK || !utf8.ValidString(name) ||
		utf8.RuneCountInString(name) < 1 || utf8.RuneCountInString(name) > 100 {
		return repository.Workspace{}, errs.NewBadRequestError(
			"Enter a workspace name of 1–100 characters and a valid retry key.", false, nil, nil, nil)
	}
	canonical, err := json.Marshal(struct {
		Name string `json:"name"`
	}{name})
	if err != nil {
		return repository.Workspace{}, workspaceFailure(err)
	}
	hash := sha256.Sum256(canonical)
	ctx, cancel := context.WithTimeout(ctx, workspaceTimeout)
	defer cancel()
	result, err := s.store.Create(ctx, actor, name, key, hash[:])
	return result, workspaceFailure(err)
}

// List reads only the actor's current workspace memberships.
func (s *WorkspaceService) List(ctx context.Context, actor uuid.UUID) ([]repository.Workspace, error) {
	ctx, cancel := context.WithTimeout(ctx, workspaceTimeout)
	defer cancel()
	result, err := s.store.List(ctx, actor)
	return result, workspaceFailure(err)
}

// Summary requires both actor and workspace scope and fresh membership.
func (s *WorkspaceService) Summary(ctx context.Context, scope repository.Scope) (repository.Workspace, error) {
	ctx, cancel := context.WithTimeout(ctx, workspaceTimeout)
	defer cancel()
	result, err := s.store.Summary(ctx, scope)
	return result, workspaceFailure(err)
}

// RequireWrite applies the closed role matrix inside the caller's locked transaction.
func (s *WorkspaceService) RequireWrite(ctx context.Context, tx pgx.Tx, scope repository.Scope) error {
	workspace, err := s.store.LockScope(ctx, tx, scope, false)
	if err != nil {
		return workspaceFailure(err)
	}
	switch workspace.Role {
	case "owner", "admin", "member":
		return nil
	default:
		return errs.NewForbiddenError("You have view-only access.", false)
	}
}

func workspaceFailure(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, repository.ErrWorkspaceConflict) {
		return &errs.HTTPError{Code: "IDEMPOTENCY_CONFLICT",
			Message: "This retry key was used for a different request.", Status: http.StatusConflict}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return errs.NewNotFoundError("Workspace not found", false, nil)
	}
	return &errs.HTTPError{Code: "SERVICE_UNAVAILABLE",
		Message: "Workspace temporarily unavailable", Status: http.StatusServiceUnavailable}
}
