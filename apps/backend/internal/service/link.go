package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"unicode/utf8"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/errs"
	"github.com/6sLOGAN78/flux/internal/repository"
	"github.com/6sLOGAN78/flux/internal/safety"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// LinkService enforces destination policy and fresh server-side capabilities.
type LinkService struct {
	store     *repository.LinkRepository
	workspace *WorkspaceService
	policy    config.LinksConfig
}

// NewLinkService injects transactional storage and validated operator policy.
func NewLinkService(store *repository.LinkRepository, workspace *WorkspaceService,
	policy config.LinksConfig,
) *LinkService {
	return &LinkService{store: store, workspace: workspace, policy: policy}
}

// ManagedHost exposes only the fixed operator-configured management hostname.
func (s *LinkService) ManagedHost() string { return s.policy.ManagedHost }

// Create validates canonical payload without DNS, HTTP or preview egress.
func (s *LinkService) Create(ctx context.Context, scope repository.Scope,
	destination, title, key string,
) (repository.Link, error) {
	canonical, err := safety.ValidateDestination(destination, s.policy)
	keyOK, matchErr := regexp.MatchString(`^[A-Za-z0-9_-]{16,128}$`, key)
	if err != nil || matchErr != nil || !keyOK || !utf8.ValidString(title) || utf8.RuneCountInString(title) > 200 {
		return repository.Link{}, errs.NewBadRequestError("Enter a public HTTP(S) destination, "+
			"a title of at most 200 characters and a valid retry key.", false, nil, nil, nil)
	}
	payload, err := json.Marshal(struct {
		Destination string `json:"destination"`
		Title       string `json:"title"`
	}{canonical, title})
	if err != nil {
		return repository.Link{}, linkFailure(err)
	}
	hash := sha256.Sum256(payload)
	ctx, cancel := context.WithTimeout(ctx, workspaceTimeout)
	defer cancel()
	result, err := s.store.Create(ctx, scope, s.policy.ManagedHost, canonical, title,
		key, hash[:], s.workspace.RequireWrite)
	return result, linkFailure(err)
}

// Detail preserves durable creator provenance without granting former access.
func (s *LinkService) Detail(ctx context.Context, scope repository.Scope, id uuid.UUID) (repository.Link, error) {
	ctx, cancel := context.WithTimeout(ctx, workspaceTimeout)
	defer cancel()
	authorize := func(ctx context.Context, tx pgx.Tx, scope repository.Scope) error {
		return s.workspace.RequireCapability(ctx, tx, scope, CapabilityRead)
	}
	result, err := s.store.Detail(ctx, scope, id, authorize)
	return result, linkFailure(err)
}

func linkFailure(err error) error {
	if err == nil {
		return nil
	}
	var typed *errs.HTTPError
	if errors.As(err, &typed) {
		return err
	}
	if errors.Is(err, repository.ErrWorkspaceConflict) {
		return &errs.HTTPError{Code: "IDEMPOTENCY_CONFLICT",
			Message: "This retry key was used for a different request.", Status: http.StatusConflict}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return errs.NewNotFoundError("Link not found", false, nil)
	}
	return errors.Join(&errs.HTTPError{Code: errs.MakeUpperCaseWithUnderscores(
		http.StatusText(http.StatusServiceUnavailable)), Message: "Links temporarily unavailable",
		Status: http.StatusServiceUnavailable}, err)
}
