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
	"github.com/go-jose/go-jose/v3"
	josejwt "github.com/go-jose/go-jose/v3/jwt"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

var browserFixture = flag.Bool("browser-fixture", false, "hold test-only loopback bearer fixture")

const fixtureIssuer = "https://fixture.clerk.accounts.dev"
const fixtureParty = "http://127.0.0.1:3100"

type signedProvider struct {
	key      *rsa.PrivateKey
	server   *httptest.Server
	mu       sync.Mutex
	status   string
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
			_ = json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{Key: &key.PublicKey, KeyID: "fixture", Algorithm: "RS256", Use: "sig"}}})
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
	claims := map[string]any{"iss": fixtureIssuer, "azp": fixtureParty, "sub": "user_fixture", "sid": "sess_fixture", "iat": time.Now().Unix(), "nbf": time.Now().Add(-time.Second).Unix(), "exp": time.Now().Add(3 * time.Minute).Unix(), "v": 2}
	for name, value := range changes {
		if value == nil {
			delete(claims, name)
		} else {
			claims[name] = value
		}
	}
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: p.key}, (&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", "fixture"))
	require.NoError(t, err)
	token, err := josejwt.Signed(signer).Claims(claims).CompactSerialize()
	require.NoError(t, err)
	return token
}

func fixtureRouter(cfg *config.Config, db *database.Database, _ *signedProvider) http.Handler {
	log := zerolog.Nop()
	srv := &server.Server{Config: cfg, Logger: &log, DB: db}
	return router.NewRouter(srv, &handler.Handlers{OpenAPI: handler.NewOpenAPIHandler(srv)}, &service.Services{})
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
			_ = json.NewEncoder(w).Encode(map[string]any{"api": api.URL, "token": token})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer protocol.Close()
	if *browserFixture {
		signalCtx, signalCancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
		defer signalCancel()
		fmt.Println("FLUX_BROWSER_FIXTURE " + protocol.URL)
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
	require.Less(t, response.StatusCode, 500)
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
		name    string
		changes map[string]any
		status  int
	}{
		{"active", nil, 200},
		{"foreign-issuer", map[string]any{"iss": "https://other.clerk.accounts.dev"}, 401},
		{"foreign-party", map[string]any{"azp": "https://evil.test"}, 401},
		{"missing-party", map[string]any{"azp": nil}, 401},
		{"empty-subject", map[string]any{"sub": ""}, 401},
		{"missing-subject", map[string]any{"sub": nil}, 401},
		{"missing-session", map[string]any{"sid": nil}, 401},
		{"session-path", map[string]any{"sid": "../jwks"}, 401},
		{"expired", map[string]any{"exp": time.Now().Add(-time.Minute).Unix()}, 401},
		{"missing-expiry", map[string]any{"exp": nil}, 401},
		{"future-not-before", map[string]any{"nbf": time.Now().Add(time.Hour).Unix()}, 401},
		{"missing-not-before", map[string]any{"nbf": nil}, 401},
		{"future-issued", map[string]any{"iat": time.Now().Add(time.Hour).Unix()}, 401},
		{"revoked", map[string]any{"sid": "sess_revoked"}, 401},
		{"pending", map[string]any{"sid": "sess_pending"}, 401},
		{"mismatch", map[string]any{"sid": "sess_mismatch"}, 401},
		{"session-not-found", map[string]any{"sid": "sess_missing"}, 401},
		{"provider-outage", map[string]any{"sid": "sess_outage"}, 503},
		{"provider-deadline", map[string]any{"sid": "sess_slow"}, 503},
		{"organization-not-authority", map[string]any{"o": map[string]any{"id": "org_other", "rol": "admin"}}, 200},
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
