package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"net/http"
	"net/mail"
	"regexp"
	"strings"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/errs"
	"github.com/6sLOGAN78/flux/internal/repository"
	"github.com/google/uuid"
)

// InvitationService queues durable encrypted intent without provider calls.
type InvitationService struct {
	store *repository.InvitationRepository
	team  *TeamService
	cfg   config.InvitationConfig
}

// NewInvitationService injects current SQL authority and external encryption policy.
func NewInvitationService(store *repository.InvitationRepository, team *TeamService,
	cfg config.InvitationConfig,
) *InvitationService {
	return &InvitationService{store: store, team: team, cfg: cfg}
}

// AllowsInvitation is the closed grant policy; owner is never invited.
func AllowsInvitation(actor, role string) bool {
	return role != roleOwner && AllowsRoleChange(actor, role, role)
}

func normalizeInvitationEmail(email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	address, err := mail.ParseAddress(email)
	// Keep the server's accepted projection within canonical ZInvitation.email.
	// This is syntactic validation only; no provider-specific aliases or DNS lookup.
	canonical, patternErr := regexp.MatchString(
		`^[a-z0-9_'+.-]*[a-z0-9_+-]@([a-z0-9][a-z0-9-]*\.)+[a-z]{2,}$`, email)
	if err != nil || patternErr != nil || !canonical || strings.HasPrefix(email, ".") ||
		strings.Contains(email, "..") || len(email) > 254 || address.Address != email {
		return "", errs.NewBadRequestError("Enter a valid email address.", false, nil, nil, nil)
	}
	return email, nil
}

// Create validates public input then commits invitation, intent, audit and ledger.
func (s *InvitationService) Create(ctx context.Context, scope repository.Scope,
	email, role, key string,
) (repository.Invitation, error) {
	email, err := normalizeInvitationEmail(email)
	if err != nil {
		return repository.Invitation{}, err
	}
	validKey, keyErr := regexp.MatchString(`^[A-Za-z0-9_-]{16,128}$`, key)
	if keyErr != nil || !validKey || !AllowsInvitation(roleOwner, role) {
		return repository.Invitation{}, errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	if s == nil || s.team == nil {
		return repository.Invitation{}, workspaceFailure(errors.New("invitation unavailable"))
	}
	if err = s.cfg.Validate(); err != nil {
		return repository.Invitation{}, workspaceFailure(err)
	}
	hash := sha256.Sum256([]byte(email + "\n" + role))
	ctx, cancel := context.WithTimeout(ctx, workspaceTimeout)
	defer cancel()
	prepare := func(invite, delivery uuid.UUID) (repository.DeliveryIntent, []byte, error) {
		return sealInvitation(s.cfg, scope, invite, delivery, email)
	}
	result, err := s.store.Create(ctx, scope, email, role, key, hash[:],
		s.team.authorizeMutation, prepare, AllowsInvitation)
	if errors.Is(err, repository.ErrInvitationPending) {
		return repository.Invitation{}, &errs.HTTPError{Code: "INVITATION_PENDING",
			Message: "An invitation is already pending for this email. Resend or revoke it.", Status: http.StatusConflict}
	}
	return result, teamMutationFailure(err)
}

// List uses a fixed 25-row page with an honest continuation position.
func (s *InvitationService) List(ctx context.Context, scope repository.Scope,
	after uuid.UUID,
) ([]repository.Invitation, *string, error) {
	if s == nil || s.team == nil {
		return nil, nil, workspaceFailure(errors.New("invitation unavailable"))
	}
	ctx, cancel := context.WithTimeout(ctx, workspaceTimeout)
	defer cancel()
	items, err := s.store.List(ctx, scope, after, s.team.authorizeMutation)
	if err != nil {
		return nil, nil, teamMutationFailure(err)
	}
	var next *string
	const size = 25
	if len(items) > size {
		items = items[:size]
		position := items[size-1].ID.String()
		next = &position
	}
	return items, next, nil
}
