package service_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/database"
	"github.com/6sLOGAN78/flux/internal/handler"
	"github.com/6sLOGAN78/flux/internal/router"
	"github.com/6sLOGAN78/flux/internal/server"
	"github.com/6sLOGAN78/flux/internal/service"
	fluxTesting "github.com/6sLOGAN78/flux/internal/testing"
	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwks"
	"github.com/clerk/clerk-sdk-go/v2/session"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/go-jose/go-jose/v3"
	josejwt "github.com/go-jose/go-jose/v3/jwt"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

//nolint:gochecknoglobals // Go parses test-binary flags before tests run; this never enters production.
var browserFixture = flag.Bool("browser-fixture", false, "hold test-only loopback bearer fixture")

const fixtureIssuer = "https://fixture.clerk.accounts.dev"
const fixtureParty = "http://127.0.0.1:3100"

type signedProvider struct {
	key      *rsa.PrivateKey
	server   *httptest.Server
	status   string
	mu       sync.Mutex
	failure  bool
	requests int
}

func newSignedProvider(t *testing.T) *signedProvider {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	p := &signedProvider{key: key, status: "active"}
	p.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		p.mu.Lock()
		defer p.mu.Unlock()
		if p.failure {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		if r.URL.Path == "/jwks" {
			_ = json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{
				Key: &key.PublicKey, KeyID: "fixture", Algorithm: "RS256", Use: "sig"}}})
			return
		}
		if strings.HasPrefix(r.URL.Path, "/sessions/") {
			p.requests++
			id := strings.TrimPrefix(r.URL.Path, "/sessions/")
			status, userID := p.status, "user_fixture"
			switch id {
			case "sess_revoked":
				status = "revoked"
			case "sess_pending":
				status = "pending"
			case "sess_mismatch":
				userID = "user_other"
			case "sess_outage":
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			case "sess_missing":
				w.WriteHeader(http.StatusNotFound)
				_, _ = io.WriteString(w, `{"errors":[{"code":"resource_not_found","message":"private provider session"}]}`)
				return
			case "sess_slow":
				<-r.Context().Done()
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"id": id, "user_id": userID, "status": status})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(p.server.Close)
	return p
}

func (p *signedProvider) token(t *testing.T, changes map[string]any) string {
	t.Helper()
	claims := map[string]any{"iss": fixtureIssuer, "azp": fixtureParty, "sub": "user_fixture", "sid": "sess_fixture",
		"iat": time.Now().Unix(), "nbf": time.Now().Add(-time.Second).Unix(),
		"exp": time.Now().Add(3 * time.Minute).Unix(), "v": 2}
	for name, value := range changes {
		if value == nil {
			delete(claims, name)
		} else {
			claims[name] = value
		}
	}
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: p.key},
		(&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", "fixture"))
	require.NoError(t, err)
	token, err := josejwt.Signed(signer).Claims(claims).CompactSerialize()
	require.NoError(t, err)
	return token
}

func fixtureRouter(cfg *config.Config, db *database.Database, p *signedProvider) http.Handler {
	log := zerolog.Nop()
	srv := &server.Server{Config: cfg, Logger: &log, DB: db}
	cfg.Auth = config.AuthConfig{SecretKey: "test-only", Issuer: fixtureIssuer, AuthorizedParties: []string{fixtureParty}}
	clients := &clerk.ClientConfig{BackendConfig: clerk.BackendConfig{
		URL: clerk.String(p.server.URL), Key: clerk.String(cfg.Auth.SecretKey),
		HTTPClient: &http.Client{Timeout: 3 * time.Second}}}
	auth := service.NewAuthServiceWithClients(cfg.Auth, service.AuthClients{
		JWKS: jwks.NewClient(clients), Sessions: session.NewClient(clients), Users: user.NewClient(clients)})
	return router.NewRouter(srv, &handler.Handlers{OpenAPI: handler.NewOpenAPIHandler(srv)}, &service.Services{Auth: auth})
}

// TestBearerBrowserFixture uses the actual router and pinned migrated PostgreSQL.
// Only an explicit test-binary flag keeps this loopback transport ready for browsers.
func TestBearerBrowserFixture(t *testing.T) {
	db, cleanup := fluxTesting.SetupTestDB(t)
	defer cleanup()
	db.Config.Server.CORSAllowedOrigins = []string{fixtureParty}
	p := newSignedProvider(t)
	api := httptest.NewServer(fixtureRouter(db.Config, &database.Database{Pool: db.Pool}, p))
	defer api.Close()
	token := p.token(t, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	stop := make(chan struct{})
	var once sync.Once
	protocol := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/stop":
			once.Do(func() { close(stop) })
			w.WriteHeader(http.StatusNoContent)
		case "/client":
			_ = json.NewEncoder(w).Encode(map[string]any{"api": api.URL, "token": token, "client": browserSessionClient(token)})
		case "/cases":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"api": api.URL, "revoked": p.token(t, map[string]any{"sid": "sess_revoked"}),
				"expired": p.token(t, map[string]any{"exp": time.Now().Add(-time.Minute).Unix()}),
				"foreign": p.token(t, map[string]any{"iss": "https://other.clerk.accounts.dev"}),
				"outage":  p.token(t, map[string]any{"sid": "sess_outage"})})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer protocol.Close()
	if *browserFixture {
		signalCtx, signalCancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
		defer signalCancel()
		_, err := fmt.Fprintln(os.Stdout, "FLUX_BROWSER_FIXTURE "+protocol.URL)
		require.NoError(t, err)
		select {
		case <-stop:
		case <-signalCtx.Done():
		}
		return
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, api.URL+"/api/v1/me", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	response, err := (&http.Client{Timeout: 3 * time.Second}).Do(req)
	require.NoError(t, err)
	defer response.Body.Close()
	_, err = io.Copy(io.Discard, response.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
}

//nolint:lll // Explicit test-only provider JSON preserves the SDK wire representation.
func browserSessionClient(token string) map[string]any {
	now := time.Now().UnixMilli()
	user := map[string]any{"object": "user", "id": "user_fixture", "username": nil, "first_name": "Local", "last_name": "Fixture", "image_url": "", "has_image": false, "primary_email_address_id": "idn_fixture", "primary_phone_number_id": nil, "primary_web3_wallet_id": nil, "password_enabled": true, "two_factor_enabled": false, "totp_enabled": false, "backup_code_enabled": false, "email_addresses": []any{map[string]any{"object": "email_address", "id": "idn_fixture", "email_address": "local@example.test", "verification": map[string]any{"status": "verified", "strategy": "email_code"}, "linked_to": []any{}}}, "phone_numbers": []any{}, "web3_wallets": []any{}, "external_accounts": []any{}, "organization_memberships": []any{}, "public_metadata": map[string]any{}, "unsafe_metadata": map[string]any{}, "created_at": now, "updated_at": now}
	session := map[string]any{"object": "session", "id": "sess_fixture", "status": "active", "expire_at": now + 600000, "abandon_at": now + 600000, "last_active_at": now, "last_active_organization_id": nil, "last_active_token": map[string]any{"object": "token", "jwt": token}, "user": user, "public_user_data": map[string]any{"first_name": "Local", "last_name": "Fixture", "identifier": "local@example.test", "user_id": "user_fixture"}, "created_at": now, "updated_at": now}
	return map[string]any{"object": "client", "id": "client_fixture", "sessions": []any{session}, "sign_in": nil, "sign_up": nil, "last_active_session_id": "sess_fixture", "cookie_expires_at": now + 600000, "created_at": now, "updated_at": now}
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
			p := newSignedProvider(t)
			cfg := &config.Config{Server: config.ServerConfig{CORSAllowedOrigins: []string{fixtureParty}}}
			api := httptest.NewServer(fixtureRouter(cfg, nil, p))
			defer api.Close()
			started := time.Now()
			status, body, cache := requestMe(t, api.URL, p.token(t, tc.changes), false)
			require.Equal(t, tc.status, status)
			require.Equal(t, "no-store", cache)
			require.Less(t, time.Since(started), 3*time.Second)
			if status == 200 {
				require.JSONEq(t, `{"authenticated":true}`, body)
			} else {
				require.NotContains(t, body, "user_fixture")
				require.NotContains(t, body, "sess_")
				require.NotContains(t, body, fixtureIssuer)
			}
		})
	}
}

func TestBearerCredentialsAndRevocation(t *testing.T) {
	p := newSignedProvider(t)
	cfg := &config.Config{Server: config.ServerConfig{CORSAllowedOrigins: []string{fixtureParty}}}
	api := httptest.NewServer(fixtureRouter(cfg, nil, p))
	defer api.Close()
	token := p.token(t, nil)
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
	p.mu.Lock()
	p.status = "revoked"
	p.mu.Unlock()
	status, _, _ = requestMe(t, api.URL, token, false)
	require.Equal(t, 401, status)
	p.mu.Lock()
	require.Equal(t, 2, p.requests)
	p.mu.Unlock()
}

func TestBearerConcurrentActiveChecks(t *testing.T) {
	p := newSignedProvider(t)
	cfg := &config.Config{Server: config.ServerConfig{CORSAllowedOrigins: []string{fixtureParty}}}
	api := httptest.NewServer(fixtureRouter(cfg, nil, p))
	defer api.Close()
	token := p.token(t, nil)
	var group sync.WaitGroup
	for range 10 {
		group.Go(func() { status, _, _ := requestMe(t, api.URL, token, false); require.Equal(t, 200, status) })
	}
	group.Wait()
	p.mu.Lock()
	defer p.mu.Unlock()
	require.Equal(t, 10, p.requests)
}
