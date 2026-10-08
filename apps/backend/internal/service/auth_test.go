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
			_ = json.NewEncoder(w).Encode(map[string]any{"id": strings.TrimPrefix(r.URL.Path, "/sessions/"), "user_id": "user_fixture", "status": p.status})
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

func fixtureRouter(cfg *config.Config, db *database.Database) http.Handler {
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
	api := httptest.NewServer(fixtureRouter(db.Config, &database.Database{Pool: db.Pool}))
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
