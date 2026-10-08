package service

import (
	"context"
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/6sLOGAN78/flux/internal/repository"
	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/jackc/pgx/v5"
)

// IdentityResolver resolves a verified actor to committed internal identity.
type IdentityResolver interface {
	Resolve(context.Context, Actor) (repository.User, error)
}

const identityTimeout = 3 * time.Second

// IdentityService fetches verified profiles only for first-time mappings.
type IdentityService struct {
	users *repository.UserRepository
	auth  *AuthService
}

// NewIdentityService injects the store and the configured provider boundary.
func NewIdentityService(users *repository.UserRepository, auth *AuthService) *IdentityService {
	return &IdentityService{users: users, auth: auth}
}

// Resolve never merges accounts by email or returns an uncommitted identity.
func (s *IdentityService) Resolve(ctx context.Context, actor Actor) (repository.User, error) {
	if s == nil || s.auth == nil || !s.auth.configured || s.auth.clients.Users == nil {
		return repository.User{}, authUnavailable()
	}
	if actor.Issuer != s.auth.issuer || !providerID(actor.Subject, "user_") {
		return repository.User{}, authDenied()
	}
	ctx, cancel := context.WithTimeout(ctx, identityTimeout)
	defer cancel()
	existing, err := s.users.Find(ctx, actor.Issuer, actor.Subject)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return repository.User{}, authUnavailable()
	}
	profileCtx, profileCancel := context.WithTimeout(ctx, authTimeout)
	defer profileCancel()
	profile, err := s.auth.clients.Users.Get(profileCtx, actor.Subject)
	if err != nil {
		return repository.User{}, authUnavailable()
	}
	email, valid := verifiedEmail(profile, actor.Subject)
	if !valid {
		return repository.User{}, authDenied()
	}
	created, err := s.users.Create(ctx, actor.Issuer, actor.Subject, email)
	if err != nil {
		return repository.User{}, authUnavailable()
	}
	return created, nil
}

func verifiedEmail(profile *clerk.User, subject string) (string, bool) {
	if profile == nil || profile.ID != subject || profile.Banned || profile.Locked ||
		profile.PrimaryEmailAddressID == nil {
		return "", false
	}
	for _, email := range profile.EmailAddresses {
		if email == nil || email.ID != *profile.PrimaryEmailAddressID || email.Verification == nil ||
			email.Verification.Status != "verified" {
			continue
		}
		value := strings.TrimSpace(email.EmailAddress)
		parsed, err := mail.ParseAddress(value)
		if err != nil || parsed.Address != value || !canonicalEmail(value) {
			return "", false
		}
		return strings.ToLower(value), true
	}
	return "", false
}

// Match the authored Zod 3 email response policy before persisting provider input.
// ParseAddress alone also accepts local-only domains and address literals.
func canonicalEmail(value string) bool {
	if len(value) > 320 || strings.HasPrefix(value, ".") || strings.Contains(value, "..") {
		return false
	}
	valid, err := regexp.MatchString(
		`^[A-Za-z0-9_'+\-.]*[A-Za-z0-9_+\-]@([A-Za-z0-9][A-Za-z0-9\-]*\.)+[A-Za-z]{2,}$`, value)
	return err == nil && valid
}
