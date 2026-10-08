// Package service collects application services and authentication configuration.
package service

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/errs"
	"github.com/6sLOGAN78/flux/internal/server"
	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwks"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/clerk/clerk-sdk-go/v2/session"
	"github.com/clerk/clerk-sdk-go/v2/user"
)

const (
	authTimeout     = 2 * time.Second
	authIdleTimeout = 30 * time.Second
)

// SessionReader is the provider lookup required on every authenticated request.
type SessionReader interface {
	Get(context.Context, string) (*clerk.Session, error)
}

// UserReader supplies verified provider profiles for subsequent durable mapping.
type UserReader interface {
	Get(context.Context, string) (*clerk.User, error)
}

// AuthClients are explicit SDK dependencies; production never uses the global key.
type AuthClients struct {
	JWKS     *jwks.Client
	Sessions SessionReader
	Users    UserReader
	Clock    clerk.Clock
}

// Actor contains verified provider identity only, never organization authority.
type Actor struct{ Issuer, Subject, SessionID string }

// AuthService verifies cryptographic claims and current provider session state.
type AuthService struct {
	clients    AuthClients
	parties    map[string]bool
	issuer     string
	configured bool
}

// NewAuthService constructs fixed-endpoint SDK clients with bounded transport.
func NewAuthService(s *server.Server) *AuthService {
	transport := &http.Transport{TLSHandshakeTimeout: authTimeout, ResponseHeaderTimeout: authTimeout,
		IdleConnTimeout: authIdleTimeout}
	transport.Proxy = nil
	client := &http.Client{Transport: transport, Timeout: authTimeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
	settings := &clerk.ClientConfig{BackendConfig: clerk.BackendConfig{
		Key: clerk.String(s.Config.Auth.SecretKey), HTTPClient: client}}
	return NewAuthServiceWithClients(s.Config.Auth, AuthClients{
		JWKS: jwks.NewClient(settings), Sessions: session.NewClient(settings),
		Users: user.NewClient(settings), Clock: clerk.NewClock(),
	})
}

// NewAuthServiceWithClients permits transport injection without any runtime bypass.
func NewAuthServiceWithClients(cfg config.AuthConfig, clients AuthClients) *AuthService {
	if clients.Clock == nil {
		clients.Clock = clerk.NewClock()
	}
	auth := &AuthService{clients: clients, issuer: cfg.Issuer, parties: make(map[string]bool)}
	issuer, err := url.Parse(cfg.Issuer)
	auth.configured = err == nil && issuer.Scheme == "https" && validOrigin(issuer) &&
		cfg.SecretKey != "" && clients.JWKS != nil && clients.Sessions != nil
	for _, party := range cfg.AuthorizedParties {
		parsed, parseErr := url.Parse(party)
		if parseErr != nil || !validOrigin(parsed) || (parsed.Scheme != "https" && parsed.Scheme != "http") {
			auth.configured = false
			continue
		}
		auth.parties[party] = true
	}
	auth.configured = auth.configured && len(auth.parties) > 0
	return auth
}

func validOrigin(parsed *url.URL) bool {
	return parsed.Host != "" && parsed.User == nil && parsed.RawQuery == "" &&
		parsed.Fragment == "" && parsed.Path == ""
}

// Authenticate accepts only an explicitly provided compact bearer token. One
// deadline covers JWKS and session lookup; upstream diagnostics remain private.
func (s *AuthService) Authenticate(ctx context.Context, token string) (Actor, error) {
	if s == nil || !s.configured {
		return Actor{}, authUnavailable()
	}
	if token == "" || len(token) > 8192 || strings.ContainsAny(token, " \t\r\n") {
		return Actor{}, authDenied()
	}
	ctx, cancel := context.WithTimeout(ctx, authTimeout)
	defer cancel()
	decoded, err := jwt.Decode(ctx, &jwt.DecodeParams{Token: token})
	if err != nil || decoded.KeyID == "" {
		return Actor{}, authDenied()
	}
	keys, err := s.clients.JWKS.Get(ctx, &jwks.GetParams{})
	if err != nil || keys == nil {
		return Actor{}, authUnavailable()
	}
	key := findKey(keys, decoded.KeyID)
	if key == nil {
		return Actor{}, authDenied()
	}
	claims, err := jwt.Verify(ctx, &jwt.VerifyParams{Token: token, JWK: key, Clock: s.clients.Clock,
		AuthorizedPartyHandler: func(party string) bool { return party != "" && s.parties[party] }})
	if err != nil || !s.validClaims(claims) {
		return Actor{}, authDenied()
	}
	active, err := s.clients.Sessions.Get(ctx, claims.SessionID)
	if err != nil {
		var providerError *clerk.APIErrorResponse
		if errors.As(err, &providerError) && providerError.HTTPStatusCode == http.StatusNotFound {
			return Actor{}, authDenied()
		}
		return Actor{}, authUnavailable()
	}
	if active == nil || active.Status != "active" || active.ID != claims.SessionID || active.UserID != claims.Subject {
		return Actor{}, authDenied()
	}
	return Actor{Issuer: claims.Issuer, Subject: claims.Subject, SessionID: claims.SessionID}, nil
}

func findKey(keys *clerk.JSONWebKeySet, keyID string) *clerk.JSONWebKey {
	for _, candidate := range keys.Keys {
		if candidate != nil && candidate.KeyID == keyID {
			return candidate
		}
	}
	return nil
}

func (s *AuthService) validClaims(claims *clerk.SessionClaims) bool {
	if claims == nil || claims.Issuer != s.issuer || !providerID(claims.Subject, "user_") ||
		!providerID(claims.SessionID, "sess_") || claims.Expiry == nil || claims.NotBefore == nil || claims.IssuedAt == nil {
		return false
	}
	now := s.clients.Clock.Now().Unix()
	return *claims.Expiry > now && *claims.NotBefore <= now && *claims.IssuedAt <= now &&
		*claims.Expiry > *claims.NotBefore && *claims.Expiry > *claims.IssuedAt
}

func providerID(value, prefix string) bool {
	if !strings.HasPrefix(value, prefix) || len(value) <= len(prefix) || len(value) > 256 {
		return false
	}
	for _, char := range value {
		switch {
		case char >= 'a' && char <= 'z', char >= 'A' && char <= 'Z', char >= '0' && char <= '9', char == '_':
		default:
			return false
		}
	}
	return true
}

func authDenied() *errs.HTTPError { return errs.NewUnauthorizedError("Authentication required", false) }
func authUnavailable() *errs.HTTPError {
	return &errs.HTTPError{Code: "SERVICE_UNAVAILABLE", Message: "Authentication temporarily unavailable",
		Status: http.StatusServiceUnavailable}
}
