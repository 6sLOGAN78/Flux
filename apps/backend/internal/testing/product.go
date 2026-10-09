//nolint:goconst,mnd // Synthetic provider wire fixtures preserve literal fields and lifetimes for browser tests.
package testing

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
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
	"github.com/go-jose/go-jose/v3"
	josejwt "github.com/go-jose/go-jose/v3/jwt"
	"github.com/stretchr/testify/require"
)

const fixtureIssuer = "https://fixture.clerk.accounts.dev"
const fixtureParty = "http://127.0.0.1:3100"

// SignedProvider is an injected, RSA-signed provider fixture for tests only.
type SignedProvider struct {
	key      *rsa.PrivateKey
	server   *httptest.Server
	status   string
	mu       sync.Mutex
	failure  bool
	requests int
}

// NewSignedProvider owns its local transport through the calling test cleanup.
func NewSignedProvider(t *testing.T) *SignedProvider {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	p := &SignedProvider{key: key, status: "active"}
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
		if r.URL.Path == "/users/user_fixture" {
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "user_fixture", "primary_email_address_id": "email_fixture",
				"email_addresses": []any{map[string]any{"id": "email_fixture", "email_address": "local@example.test",
					"verification": map[string]any{"status": "verified"}}}})
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

// Token signs short-lived claims for the actual SDK verification path.
func (p *SignedProvider) Token(t *testing.T, changes map[string]any) string {
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

// URL is the injected test-only provider transport.
func (p *SignedProvider) URL() string { return p.server.URL }

// SetStatus changes the provider session under its synchronization boundary.
func (p *SignedProvider) SetStatus(status string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.status = status
}

// SessionRequests returns the observed active-session checks.
func (p *SignedProvider) SessionRequests() int { p.mu.Lock(); defer p.mu.Unlock(); return p.requests }

// RunBrowserProductFixture owns migrated PostgreSQL and the existing bounded
// loopback protocol. Application composition is supplied by external test packages.
func RunBrowserProductFixture(t *testing.T, hold bool,
	factory func(*config.Config, *database.Database, *SignedProvider) http.Handler,
) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()
	db.Config.Server.CORSAllowedOrigins = []string{fixtureParty}
	p := NewSignedProvider(t)
	api := httptest.NewServer(factory(db.Config, &database.Database{Pool: db.Pool}, p))
	defer api.Close()
	token := p.Token(t, nil)
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
			_ = json.NewEncoder(w).Encode(map[string]any{"api": api.URL, "token": token, "client": BrowserSessionClient(token)})
		case "/restore-reset", "/restore-remove":
			// Local harness only: mutate real PostgreSQL below application layers.
			// No test route, key or authorization bypass enters a production router.
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			var err error
			if r.URL.Path == "/restore-reset" {
				_, err = db.Pool.Exec(r.Context(), "DELETE FROM workspace_preferences WHERE user_id IN "+
					"(SELECT id FROM users WHERE issuer=$1 AND subject=$2)", fixtureIssuer, "user_fixture")
			} else {
				_, err = db.Pool.Exec(r.Context(), "DELETE FROM memberships WHERE workspace_id::text=$1 AND user_id IN "+
					"(SELECT id FROM users WHERE issuer=$2 AND subject=$3)",
					r.URL.Query().Get("workspace"), fixtureIssuer, "user_fixture")
			}
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		case "/cases":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"api": api.URL, "revoked": p.Token(t, map[string]any{"sid": "sess_revoked"}),
				"expired": p.Token(t, map[string]any{"exp": time.Now().Add(-time.Minute).Unix()}),
				"foreign": p.Token(t, map[string]any{"iss": "https://other.clerk.accounts.dev"}),
				"outage":  p.Token(t, map[string]any{"sid": "sess_outage"})})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer protocol.Close()
	if hold {
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
	defer func() { require.NoError(t, response.Body.Close()) }()
	_, err = io.Copy(io.Discard, response.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
}

// BrowserSessionClient preserves the SDK session wire format without live factors.
//
//nolint:lll // Explicit test-only provider JSON preserves the SDK wire representation.
func BrowserSessionClient(token string) map[string]any {
	now := time.Now().UnixMilli()
	user := map[string]any{"object": "user", "id": "user_fixture", "username": nil, "first_name": "Local", "last_name": "Fixture", "image_url": "", "has_image": false, "primary_email_address_id": "idn_fixture", "primary_phone_number_id": nil, "primary_web3_wallet_id": nil, "password_enabled": true, "two_factor_enabled": false, "totp_enabled": false, "backup_code_enabled": false, "email_addresses": []any{map[string]any{"object": "email_address", "id": "idn_fixture", "email_address": "local@example.test", "verification": map[string]any{"status": "verified", "strategy": "email_code"}, "linked_to": []any{}}}, "phone_numbers": []any{}, "web3_wallets": []any{}, "external_accounts": []any{}, "organization_memberships": []any{}, "public_metadata": map[string]any{}, "unsafe_metadata": map[string]any{}, "created_at": now, "updated_at": now}
	session := map[string]any{"object": "session", "id": "sess_fixture", "status": "active", "expire_at": now + 600000, "abandon_at": now + 600000, "last_active_at": now, "last_active_organization_id": nil, "last_active_token": map[string]any{"object": "token", "jwt": token}, "user": user, "public_user_data": map[string]any{"first_name": "Local", "last_name": "Fixture", "identifier": "local@example.test", "user_id": "user_fixture"}, "created_at": now, "updated_at": now}
	return map[string]any{"object": "client", "id": "client_fixture", "sessions": []any{session}, "sign_in": nil, "sign_up": nil, "last_active_session_id": "sess_fixture", "cookie_expires_at": now + 600000, "created_at": now, "updated_at": now}
}
