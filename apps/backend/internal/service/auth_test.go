package service_test

import (
	"context"
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/database"
	"github.com/6sLOGAN78/flux/internal/handler"
	"github.com/6sLOGAN78/flux/internal/repository"
	"github.com/6sLOGAN78/flux/internal/router"
	"github.com/6sLOGAN78/flux/internal/server"
	"github.com/6sLOGAN78/flux/internal/service"
	fluxTesting "github.com/6sLOGAN78/flux/internal/testing"
	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwks"
	"github.com/clerk/clerk-sdk-go/v2/session"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

//nolint:gochecknoglobals // Go parses test-binary flags before tests run; this never enters production.
var browserFixture = flag.Bool("browser-fixture", false, "hold test-only loopback bearer fixture")

const fixtureIssuer = "https://fixture.clerk.accounts.dev"
const fixtureParty = "http://127.0.0.1:3100"

func fixtureRouter(cfg *config.Config, db *database.Database, p *fluxTesting.SignedProvider) http.Handler {
	log := zerolog.Nop()
	srv := &server.Server{Config: cfg, Logger: &log, DB: db}
	cfg.Auth = config.AuthConfig{SecretKey: "test-only", Issuer: fixtureIssuer, AuthorizedParties: []string{fixtureParty}}
	clients := &clerk.ClientConfig{BackendConfig: clerk.BackendConfig{
		URL: clerk.String(p.URL()), Key: clerk.String(cfg.Auth.SecretKey),
		HTTPClient: &http.Client{Timeout: 3 * time.Second}}}
	auth := service.NewAuthServiceWithClients(cfg.Auth, service.AuthClients{
		JWKS: jwks.NewClient(clients), Sessions: session.NewClient(clients), Users: user.NewClient(clients)})
	var identity service.IdentityResolver = authCorpusIdentity{}
	if db != nil {
		identity = service.NewIdentityService(repository.NewUserRepository(db.Pool), auth)
	}
	return router.NewRouter(srv, &handler.Handlers{OpenAPI: handler.NewOpenAPIHandler(srv)},
		&service.Services{Auth: auth, Identity: identity})
}

// Only the provider-boundary unit corpus uses this explicit resolver. The browser
// fixture and registered product integration dispatcher use migrated PostgreSQL.
type authCorpusIdentity struct{}

func (authCorpusIdentity) Resolve(context.Context, service.Actor) (repository.User, error) {
	return repository.User{ID: uuid.MustParse("00000000-0000-4000-8000-000000000001"), Email: "local@example.test"}, nil
}

// TestBearerBrowserFixture uses the actual router and pinned migrated PostgreSQL.
// Only an explicit test-binary flag keeps this loopback transport ready for browsers.
func TestBearerBrowserFixture(t *testing.T) {
	fluxTesting.RunBrowserProductFixture(t, *browserFixture, fixtureRouter)
}

func requestMe(t *testing.T, api string, token string, cookie bool) (int, string, string) {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, api+"/api/v1/me", nil)
	require.NoError(t, err)
	if cookie {
		req.AddCookie(&http.Cookie{Name: "__session", Value: token})
	} else if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	response, err := (&http.Client{Timeout: 4 * time.Second}).Do(req)
	require.NoError(t, err)
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 65536))
	require.NoError(t, err)
	return response.StatusCode, string(body), response.Header.Get("Cache-Control")
}

func TestBearerSignedHTTPCorpus(t *testing.T) {
	cases := []struct {
		changes map[string]any
		name    string
		status  int
	}{
		{nil, "active", 200},
		{map[string]any{"iss": "https://other.clerk.accounts.dev"}, "foreign-issuer", 401},
		{map[string]any{"azp": "https://evil.test"}, "foreign-party", 401},
		{map[string]any{"azp": nil}, "missing-party", 401},
		{map[string]any{"sub": ""}, "empty-subject", 401},
		{map[string]any{"sub": nil}, "missing-subject", 401},
		{map[string]any{"sid": nil}, "missing-session", 401},
		{map[string]any{"sid": "../jwks"}, "session-path", 401},
		{map[string]any{"exp": time.Now().Add(-time.Minute).Unix()}, "expired", 401},
		{map[string]any{"exp": nil}, "missing-expiry", 401},
		{map[string]any{"nbf": time.Now().Add(time.Hour).Unix()}, "future-not-before", 401},
		{map[string]any{"nbf": nil}, "missing-not-before", 401},
		{map[string]any{"iat": time.Now().Add(time.Hour).Unix()}, "future-issued", 401},
		{map[string]any{"sid": "sess_revoked"}, "revoked", 401},
		{map[string]any{"sid": "sess_pending"}, "pending", 401},
		{map[string]any{"sid": "sess_mismatch"}, "mismatch", 401},
		{map[string]any{"sid": "sess_missing"}, "session-not-found", 401},
		{map[string]any{"sid": "sess_outage"}, "provider-outage", 503},
		{map[string]any{"sid": "sess_slow"}, "provider-deadline", 503},
		{map[string]any{"o": map[string]any{"id": "org_other", "rol": "admin"}}, "organization-not-authority", 200},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := fluxTesting.NewSignedProvider(t)
			cfg := &config.Config{Server: config.ServerConfig{CORSAllowedOrigins: []string{fixtureParty}}}
			api := httptest.NewServer(fixtureRouter(cfg, nil, p))
			defer api.Close()
			started := time.Now()
			status, body, cache := requestMe(t, api.URL, p.Token(t, tc.changes), false)
			require.Equal(t, tc.status, status)
			require.Equal(t, "no-store", cache)
			require.Less(t, time.Since(started), 3*time.Second)
			if status == 200 {
				require.JSONEq(t, `{"authenticated":true,"user":{"id":"00000000-0000-4000-8000-000000000001",`+
					`"email":"local@example.test"}}`, body)
			} else {
				require.NotContains(t, body, "user_fixture")
				require.NotContains(t, body, "sess_")
				require.NotContains(t, body, fixtureIssuer)
			}
		})
	}
}

func TestBearerCredentialsAndRevocation(t *testing.T) {
	p := fluxTesting.NewSignedProvider(t)
	cfg := &config.Config{Server: config.ServerConfig{CORSAllowedOrigins: []string{fixtureParty}}}
	api := httptest.NewServer(fixtureRouter(cfg, nil, p))
	defer api.Close()
	token := p.Token(t, nil)
	for _, tc := range []struct {
		name, token string
		cookie      bool
	}{
		{"missing", "", false}, {"cookie-only", token, true}, {"malformed", "broken", false},
		{"signature", token[:len(token)-10] + "tamperedxx", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			status, _, cache := requestMe(t, api.URL, tc.token, tc.cookie)
			require.Equal(t, 401, status)
			require.Equal(t, "no-store", cache)
		})
	}
	status, _, _ := requestMe(t, api.URL, token, false)
	require.Equal(t, 200, status)
	p.SetStatus("revoked")
	status, _, _ = requestMe(t, api.URL, token, false)
	require.Equal(t, 401, status)
	require.Equal(t, 2, p.SessionRequests())
}

func TestBearerConcurrentActiveChecks(t *testing.T) {
	p := fluxTesting.NewSignedProvider(t)
	cfg := &config.Config{Server: config.ServerConfig{CORSAllowedOrigins: []string{fixtureParty}}}
	api := httptest.NewServer(fixtureRouter(cfg, nil, p))
	defer api.Close()
	token := p.Token(t, nil)
	var group sync.WaitGroup
	for range 10 {
		group.Go(func() { status, _, _ := requestMe(t, api.URL, token, false); require.Equal(t, 200, status) })
	}
	group.Wait()
	require.Equal(t, 10, p.SessionRequests())
}
