package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
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
	cursorKey []byte
}

// NewLinkService injects transactional storage and validated operator policy.
func NewLinkService(store *repository.LinkRepository, workspace *WorkspaceService,
	policy config.LinksConfig,
) *LinkService {
	key, err := policy.CursorSigningKey()
	if err != nil {
		// API startup validates this policy. Direct callers still fail closed in List.
		return &LinkService{store: store, workspace: workspace, policy: policy}
	}
	return &LinkService{store: store, workspace: workspace, policy: policy, cursorKey: key}
}

// ManagedHost exposes only the fixed operator-configured management hostname.
func (s *LinkService) ManagedHost() string { return s.policy.ManagedHost }

// Create validates canonical payload without DNS, HTTP or preview egress.
func (s *LinkService) Create(ctx context.Context, scope repository.Scope,
	destination, title, key, customKey string,
) (repository.Link, error) {
	canonical, err := safety.ValidateDestination(destination, s.policy)
	customKey, customOK := normalizeCustomKey(customKey)
	keyOK, matchErr := regexp.MatchString(`^[A-Za-z0-9_-]{16,128}$`, key)
	if !customOK {
		return repository.Link{}, &errs.HTTPError{Code: "INVALID_CUSTOM_KEY", Status: http.StatusBadRequest,
			Message: "Enter a valid custom short key."}
	}
	if err != nil || matchErr != nil || !keyOK || !utf8.ValidString(title) || utf8.RuneCountInString(title) > 200 {
		return repository.Link{}, errs.NewBadRequestError("Enter a public HTTP(S) destination, "+
			"a title of at most 200 characters and a valid retry key.", false, nil, nil, nil)
	}
	payload, err := json.Marshal(struct {
		Destination string `json:"destination"`
		Title       string `json:"title"`
		CustomKey   string `json:"customKey,omitempty"`
	}{canonical, title, customKey})
	if err != nil {
		return repository.Link{}, linkFailure(err)
	}
	hash := sha256.Sum256(payload)
	ctx, cancel := context.WithTimeout(ctx, workspaceTimeout)
	defer cancel()
	result, err := s.store.Create(ctx, scope, s.policy.ManagedHost, canonical, title,
		key, customKey, hash[:], s.workspace.RequireWrite)
	return result, linkFailure(err)
}

func normalizeCustomKey(key string) (string, bool) {
	if key == "" {
		return "", true
	}
	// Validate ASCII before case folding: Unicode lookalikes must never become keys.
	if ok, _ := regexp.MatchString(`^[A-Za-z0-9][A-Za-z0-9_-]{2,63}$`, key); !ok {
		return "", false
	}
	key = strings.ToLower(key)
	for _, reserved := range strings.Fields("api docs live ready static login register dashboard settings links admin") {
		if key == reserved {
			return "", false
		}
	}
	return key, true
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
		return &errs.HTTPError{Code: "REQUEST_REUSE_CONFLICT",
			Message: "This retry key was used for a different request.", Status: http.StatusConflict}
	}
	if errors.Is(err, repository.ErrLinkKeyUnavailable) {
		return &errs.HTTPError{Code: "KEY_UNAVAILABLE", Status: http.StatusConflict,
			Message: "This short key is unavailable. Choose another key or generate one."}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return errs.NewNotFoundError("Link not found", false, nil)
	}
	return errors.Join(&errs.HTTPError{Code: errs.MakeUpperCaseWithUnderscores(
		http.StatusText(http.StatusServiceUnavailable)), Message: "Links temporarily unavailable",
		Status: http.StatusServiceUnavailable}, err)
}

// List freshly authorizes caller scope and seeks using a verified query-bound position.
func (s *LinkService) List(ctx context.Context, scope repository.Scope, limit int,
	cursor string,
) ([]repository.Link, *string, error) {
	if limit < 1 || limit > 100 {
		return nil, nil, errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	if len(s.cursorKey) != sha256.Size {
		return nil, nil, linkFailure(errors.New("cursor signing unavailable"))
	}
	filters := CursorFilters{State: "nondeleted"}
	var position *repository.LinkPosition
	var err error
	if cursor != "" {
		position, err = decodeCursor(s.cursorKey, cursor, scope.WorkspaceID, filters)
		if err != nil {
			return nil, nil, err
		}
	}
	ctx, cancel := context.WithTimeout(ctx, workspaceTimeout)
	defer cancel()
	authorize := func(ctx context.Context, tx pgx.Tx, scope repository.Scope) error {
		return s.workspace.RequireCapability(ctx, tx, scope, CapabilityRead)
	}
	items, err := s.store.List(ctx, scope, limit+1, position, authorize)
	if err != nil {
		return nil, nil, linkFailure(err)
	}
	var next *string
	if len(items) > limit {
		items = items[:limit]
		last := items[len(items)-1]
		token, signErr := encodeCursor(s.cursorKey, scope.WorkspaceID, filters,
			repository.LinkPosition{CreatedAt: last.CreatedAt, ID: last.ID})
		if signErr != nil {
			return nil, nil, linkFailure(signErr)
		}
		next = &token
	}
	return items, next, nil
}
