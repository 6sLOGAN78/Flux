//nolint:lll // Keep test-only provider wire data and rollback SQL readable without changing their representation.
package handler_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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
	"github.com/go-jose/go-jose/v3"
	josejwt "github.com/go-jose/go-jose/v3/jwt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

const fixtureIssuer = "https://fixture.clerk.accounts.dev"
const fixtureParty = "http://127.0.0.1:3100"

type signedProvider struct {
	key         *rsa.PrivateKey
	server      *httptest.Server
	status      string
	profileMode string
	issuer      string
	logs        bytes.Buffer
	mu          sync.Mutex
	requests    int
	profiles    int
	failure     bool
}

func newSignedProvider(t *testing.T) *signedProvider {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	p := &signedProvider{key: key, status: "active", issuer: fixtureIssuer}
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
		if strings.HasPrefix(r.URL.Path, "/users/") {
			id := strings.TrimPrefix(r.URL.Path, "/users/")
			p.profiles++
			verified := "verified"
			primary := "email_fixture"
			email := "same@example.test"
			banned := false
			switch p.profileMode {
			case "failure":
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = io.WriteString(w, `{"errors":[{"message":"PRIVATE-PROFILE-MARKER"}]}`)
				return
			case "slow":
				<-r.Context().Done()
				return
			case "mismatch":
				id = "user_unrelated"
			case "unverified":
				verified = "unverified"
			case "missingprimary":
				primary = "missing"
			case "invalidemail":
				email = "not an email"
			case "incompatibleemail":
				email = "local@localhost"
			case "banned":
				banned = true
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"id": id, "primary_email_address_id": primary, "banned": banned,
				"email_addresses": []any{map[string]any{"id": "email_fixture", "email_address": email,
					"verification": map[string]any{"status": verified}}}})
			return
		}
		if strings.HasPrefix(r.URL.Path, "/sessions/") {
			p.requests++
			id := strings.TrimPrefix(r.URL.Path, "/sessions/")
			status, userID := p.status, "user_fixture"
			if strings.HasPrefix(id, "sess_user_") {
				userID = strings.TrimPrefix(id, "sess_")
			}
			if id == "sess_other" {
				userID = "user_other"
			}
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
	claims := map[string]any{"iss": p.issuer, "azp": fixtureParty, "sub": "user_fixture", "sid": "sess_fixture",
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

func productRouter(cfg *config.Config, db *database.Database, p *signedProvider) http.Handler {
	log := zerolog.New(zerolog.SyncWriter(&p.logs))
	srv := &server.Server{Config: cfg, Logger: &log, DB: db}
	cfg.Auth = config.AuthConfig{SecretKey: "test-only", Issuer: p.issuer, AuthorizedParties: []string{fixtureParty}}
	clients := &clerk.ClientConfig{BackendConfig: clerk.BackendConfig{
		URL: clerk.String(p.server.URL), Key: clerk.String(cfg.Auth.SecretKey),
		HTTPClient: &http.Client{Timeout: 3 * time.Second}}}
	auth := service.NewAuthServiceWithClients(cfg.Auth, service.AuthClients{
		JWKS: jwks.NewClient(clients), Sessions: session.NewClient(clients), Users: user.NewClient(clients)})
	return router.NewRouter(srv, &handler.Handlers{OpenAPI: handler.NewOpenAPIHandler(srv)}, &service.Services{Auth: auth,
		Identity: service.NewIdentityService(repository.NewUserRepository(db.Pool), auth)})
}

func requestMe(t *testing.T, api string, token string, _ bool) (int, string, string) {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, api+"/api/v1/me", nil)
	require.NoError(t, err)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	response, err := (&http.Client{Timeout: 4 * time.Second}).Do(req)
	require.NoError(t, err)
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 65536))
	require.NoError(t, err)
	return response.StatusCode, string(body), response.Header.Get("Cache-Control")
}

// TestProductActualHTTP is the registered real-PostgreSQL product dispatcher.
func TestProductActualHTTP(t *testing.T) {
	db, cleanup := fluxTesting.SetupTestDB(t)
	defer cleanup()
	p := newSignedProvider(t)
	db.Config.Server.CORSAllowedOrigins = []string{fixtureParty}
	api := httptest.NewServer(productRouter(db.Config, &database.Database{Pool: db.Pool}, p))
	defer api.Close()
	token := p.token(t, nil)
	t.Run("session-mutation-boundary", func(t *testing.T) {
		cfg := *db.Config
		cfg.Server.CORSAllowedOrigins = []string{fixtureParty}
		e := productRouter(&cfg, &database.Database{Pool: db.Pool}, p).(*echo.Echo)
		checkSessionMutations(t, e, token)
	})
	t.Run("session-recovery-headers", func(t *testing.T) {
		e := productRouter(db.Config, &database.Database{Pool: db.Pool}, p).(*echo.Echo)
		checkSessionRecoveryHeaders(t, e, p.token(t, map[string]any{"sid": "sess_outage"}))
	})
	t.Run("identity-stable-committed-UUID", func(t *testing.T) {
		status, body, cache := requestMe(t, api.URL, token, false)
		require.Equal(t, 200, status)
		require.Equal(t, "no-store", cache)
		var identity struct {
			User struct {
				ID string `json:"id"`
			} `json:"user"`
		}
		require.NoError(t, json.Unmarshal([]byte(body), &identity))
		id, err := uuid.Parse(identity.User.ID)
		require.NoError(t, err, "authenticated session must resolve to a durable internal UUID")
		require.NotEqual(t, uuid.Nil, id)
		status, repeated, _ := requestMe(t, api.URL, token, false)
		require.Equal(t, 200, status)
		require.JSONEq(t, body, repeated)
		var stored string
		require.NoError(t, db.Pool.QueryRow(context.Background(),
			"SELECT id::text FROM users WHERE issuer=$1 AND subject=$2", fixtureIssuer, "user_fixture").Scan(&stored))
		require.Equal(t, identity.User.ID, stored)
		p.mu.Lock()
		require.Equal(t, 1, p.profiles, "existing mapping should not fetch profile again")
		p.mu.Unlock()
	})
	t.Run("identity-concurrent-first-mapping", func(t *testing.T) {
		other := p.token(t, map[string]any{"sub": "user_other", "sid": "sess_other"})
		var group sync.WaitGroup
		identities := make(chan string, 8)
		for range 8 {
			group.Go(func() {
				status, body, cache := requestMe(t, api.URL, other, false)
				require.Equal(t, 200, status)
				require.Equal(t, "no-store", cache)
				identities <- body
			})
		}
		group.Wait()
		close(identities)
		var first string
		for body := range identities {
			if first == "" {
				first = body
			}
			require.JSONEq(t, first, body)
		}
		var rows int
		require.NoError(t, db.Pool.QueryRow(context.Background(), "SELECT count(*) FROM users").Scan(&rows))
		require.Equal(t, 2, rows, "same verified email must never merge two provider subjects")
		var response struct {
			User struct {
				ID string `json:"id"`
			} `json:"user"`
		}
		require.NoError(t, json.Unmarshal([]byte(first), &response))
		var stored, original string
		require.NoError(t, db.Pool.QueryRow(context.Background(), "SELECT id::text FROM users WHERE issuer=$1 AND subject=$2", fixtureIssuer, "user_other").Scan(&stored))
		require.NoError(t, db.Pool.QueryRow(context.Background(), "SELECT id::text FROM users WHERE issuer=$1 AND subject=$2", fixtureIssuer, "user_fixture").Scan(&original))
		require.Equal(t, stored, response.User.ID)
		require.NotEqual(t, original, response.User.ID)
	})
	for _, tc := range []struct {
		mode   string
		status int
	}{
		{"failure", 503}, {"slow", 503}, {"mismatch", 401}, {"unverified", 401},
		{"missingprimary", 401}, {"invalidemail", 401}, {"banned", 401},
		{"incompatibleemail", 401},
	} {
		t.Run("identity-profile-"+tc.mode, func(t *testing.T) {
			provider := newSignedProvider(t)
			provider.profileMode = tc.mode
			profileAPI := httptest.NewServer(productRouter(db.Config, &database.Database{Pool: db.Pool}, provider))
			defer profileAPI.Close()
			subject := "user_" + tc.mode
			started := time.Now()
			status, body, cache := requestMe(t, profileAPI.URL, provider.token(t, map[string]any{"sub": subject, "sid": "sess_" + subject}), false)
			require.Equal(t, tc.status, status)
			require.Equal(t, "no-store", cache)
			require.Less(t, time.Since(started), 3*time.Second)
			for _, secret := range []string{subject, fixtureIssuer, "same@example.test", "PRIVATE-PROFILE-MARKER"} {
				require.NotContains(t, body, secret)
				require.NotContains(t, provider.logs.String(), secret)
			}
			var rows int
			require.NoError(t, db.Pool.QueryRow(context.Background(), "SELECT count(*) FROM users WHERE subject=$1", subject).Scan(&rows))
			require.Zero(t, rows)
		})
	}
	t.Run("identity-commit-failure-rolls-back", func(t *testing.T) {
		_, err := db.Pool.Exec(context.Background(), `CREATE FUNCTION reject_identity_commit() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'PRIVATE-COMMIT-MARKER'; END $$;
		CREATE CONSTRAINT TRIGGER reject_identity_commit AFTER INSERT ON users DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION reject_identity_commit()`)
		require.NoError(t, err)
		defer func() {
			_, dropErr := db.Pool.Exec(context.Background(), "DROP TRIGGER reject_identity_commit ON users; DROP FUNCTION reject_identity_commit()")
			require.NoError(t, dropErr)
		}()
		status, body, cache := requestMe(t, api.URL, p.token(t, map[string]any{"sub": "user_rollback", "sid": "sess_user_rollback"}), false)
		require.Equal(t, 503, status)
		require.Equal(t, "no-store", cache)
		require.NotContains(t, body, "PRIVATE-COMMIT-MARKER")
		require.NotContains(t, p.logs.String(), "PRIVATE-COMMIT-MARKER")
		var rows int
		require.NoError(t, db.Pool.QueryRow(context.Background(), "SELECT count(*) FROM users WHERE subject=$1", "user_rollback").Scan(&rows))
		require.Zero(t, rows)
	})
	t.Run("identity-configured-issuer-namespace", func(t *testing.T) {
		provider := newSignedProvider(t)
		provider.issuer = "https://second.clerk.accounts.dev"
		cfg := &config.Config{Server: config.ServerConfig{CORSAllowedOrigins: []string{fixtureParty}}}
		issuerAPI := httptest.NewServer(productRouter(cfg, &database.Database{Pool: db.Pool}, provider))
		defer issuerAPI.Close()
		status, _, _ := requestMe(t, issuerAPI.URL, provider.token(t, nil), false)
		require.Equal(t, 200, status)
		var rows int
		require.NoError(t, db.Pool.QueryRow(context.Background(), "SELECT count(DISTINCT id) FROM users WHERE subject=$1", "user_fixture").Scan(&rows))
		require.Equal(t, 2, rows, "same subject under a different configured issuer is distinct")
	})
	t.Run("identity-namespace-and-UUID-immutable", func(t *testing.T) {
		for _, statement := range []string{
			"UPDATE users SET issuer='https://changed.test' WHERE subject='user_fixture'",
			"UPDATE users SET subject='user_changed' WHERE subject='user_fixture'",
			"UPDATE users SET id=gen_random_uuid() WHERE subject='user_fixture'",
		} {
			_, err := db.Pool.Exec(context.Background(), statement)
			require.Error(t, err)
		}
	})
	t.Run("identity-request-cancellation-rolls-back", func(t *testing.T) {
		lock, err := db.Pool.Begin(context.Background())
		require.NoError(t, err)
		_, err = lock.Exec(context.Background(), "LOCK TABLE users IN SHARE MODE")
		require.NoError(t, err)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, api.URL+"/api/v1/me", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+p.token(t, map[string]any{"sub": "user_cancelled", "sid": "sess_user_cancelled"}))
		finished := make(chan error, 1)
		go func() {
			response, requestErr := (&http.Client{Timeout: 4 * time.Second}).Do(req)
			if response != nil {
				_ = response.Body.Close()
			}
			finished <- requestErr
		}()
		require.Eventually(t, func() bool {
			var waiting int
			return db.Pool.QueryRow(context.Background(), "SELECT count(*) FROM pg_stat_activity "+
				"WHERE datname=current_database() AND wait_event_type='Lock' AND query LIKE 'INSERT INTO users%'").
				Scan(&waiting) == nil && waiting == 1
		}, 2*time.Second, 20*time.Millisecond, "signed request must reach the blocked PostgreSQL insert")
		cancel()
		require.ErrorIs(t, <-finished, context.Canceled)
		require.NoError(t, lock.Rollback(context.Background()))
		require.Eventually(t, func() bool {
			var transactions int
			return db.Pool.QueryRow(context.Background(), "SELECT count(*) FROM pg_stat_activity WHERE datname=current_database() AND state='idle in transaction'").Scan(&transactions) == nil && transactions == 0
		}, 2*time.Second, 20*time.Millisecond)
		var rows int
		require.NoError(t, db.Pool.QueryRow(context.Background(), "SELECT count(*) FROM users WHERE subject=$1", "user_cancelled").Scan(&rows))
		require.Zero(t, rows)
	})
	t.Run("identity-success-logs-exclude-secrets", func(t *testing.T) {
		for _, secret := range []string{token, "same@example.test", fixtureIssuer, "user_fixture", "sess_fixture"} {
			require.NotContains(t, p.logs.String(), secret)
		}
	})
	t.Run("workspace-bootstrap", func(t *testing.T) {
		checkWorkspaceBootstrap(t, api.URL, db, p, token)
	})
}

func workspaceRequest(t *testing.T, api, token, method, path, key, body string) (int, map[string]any) {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), method, api+"/api/v1"+path, strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Origin", fixtureParty)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", key)
	response, err := (&http.Client{Timeout: 5 * time.Second}).Do(req)
	require.NoError(t, err)
	defer response.Body.Close()
	require.Equal(t, "no-store", response.Header.Get("Cache-Control"))
	var result map[string]any
	require.NoError(t, json.NewDecoder(response.Body).Decode(&result))
	return response.StatusCode, result
}

func checkWorkspaceBootstrap(t *testing.T, api string, db *fluxTesting.TestDB, p *signedProvider, token string) {
	t.Helper()
	status, first := workspaceRequest(t, api, token, "POST", "/workspaces", "workspace-bootstrap-0001", `{"name":"  Growth 🚀  "}`)
	require.Equal(t, 201, status)
	workspace := first["workspace"].(map[string]any)
	id := workspace["id"].(string)
	require.Equal(t, "Growth 🚀", workspace["name"])
	require.Equal(t, "owner", workspace["role"])
	status, repeated := workspaceRequest(t, api, token, "POST", "/workspaces", "workspace-bootstrap-0001", `{"name":"Growth 🚀"}`)
	require.Equal(t, 201, status)
	require.Equal(t, first, repeated)
	status, _ = workspaceRequest(t, api, token, "POST", "/workspaces", "workspace-bootstrap-0001", `{"name":"Changed"}`)
	require.Equal(t, 409, status)
	for _, body := range []string{`{"name":" "}`, `{"name":"` + strings.Repeat("界", 101) + `"}`, `{"name":"ok","extra":true}`} {
		status, _ = workspaceRequest(t, api, token, "POST", "/workspaces", "workspace-invalid-0001", body)
		require.Equal(t, 400, status)
	}
	status, listed := workspaceRequest(t, api, token, "GET", "/workspaces", "", "")
	require.Equal(t, 200, status)
	require.Len(t, listed["workspaces"], 1)
	status, summary := workspaceRequest(t, api, token, "GET", "/workspaces/"+id, "", "")
	require.Equal(t, 200, status)
	require.Equal(t, first, summary)
	other := p.token(t, map[string]any{"sub": "user_other", "sid": "sess_other"})
	status, denied := workspaceRequest(t, api, other, "GET", "/workspaces/"+id, "", "")
	require.Equal(t, 404, status)
	require.NotContains(t, denied, "workspace")
	var group sync.WaitGroup
	for range 4 {
		group.Go(func() {
			status, result := workspaceRequest(t, api, token, "POST", "/workspaces", "workspace-concurrent-0001", `{"name":"Concurrent"}`)
			require.Equal(t, 201, status)
			require.Equal(t, "Concurrent", result["workspace"].(map[string]any)["name"])
		})
	}
	group.Wait()
	var rows int
	require.NoError(t, db.Pool.QueryRow(context.Background(), "SELECT count(*) FROM workspaces WHERE name=$1", "Concurrent").Scan(&rows))
	require.Equal(t, 1, rows)
}

func checkSessionMutations(t *testing.T, e *echo.Echo, token string) {
	t.Helper()
	writes := 0
	verifiedActor := false
	legacyAuthority := false
	// Only the test router owns this probe; no production mutation is invented.
	e.POST("/api/v1/session-probe", func(c echo.Context) error {
		_, ok := c.Get("actor").(service.Actor)
		verifiedActor = ok
		legacyAuthority = c.Get("user_role") != nil || c.Get("permissions") != nil
		writes++
		return c.NoContent(204)
	})
	for _, item := range []struct {
		name, origin, content, bearer, body string
		status                              int
	}{
		{"missing-origin", "", "application/json", "Bearer " + token, "{}", 403},
		{"null-origin", "null", "application/json", "Bearer " + token, "{}", 403},
		{"foreign-origin", "https://foreign.test", "application/json", "Bearer " + token, "{}", 403},
		{"cross-site-form", fixtureParty, "application/x-www-form-urlencoded", "Bearer " + token, "x=y", 415},
		{"cookie-only", fixtureParty, "application/json", "", "{}", 401},
		{"malformed-header", fixtureParty, "application/json", "Bearer  " + token, "{}", 401},
		{"body-limit", fixtureParty, "application/json", "Bearer " + token, strings.Repeat("x", 65537), 413},
		{"allowed", fixtureParty, "application/json; charset=utf-8", "Bearer " + token, "{}", 204},
	} {
		t.Run(item.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/session-probe", strings.NewReader(item.body))
			req.Header.Set("Origin", item.origin)
			req.Header.Set("Content-Type", item.content)
			if item.bearer != "" {
				req.Header.Set("Authorization", item.bearer)
			}
			req.Header.Set("Cookie", "__session=ambient-cookie")
			response := httptest.NewRecorder()
			e.ServeHTTP(response, req)
			require.Equal(t, item.status, response.Code)
			require.Equal(t, "no-store", response.Header().Get("Cache-Control"))
			require.NotEmpty(t, response.Header().Get("X-Request-ID"))
			require.NotContains(t, response.Body.String(), token)
		})
	}
	require.Equal(t, 1, writes)
	require.True(t, verifiedActor)
	require.False(t, legacyAuthority)
}

func checkSessionRecoveryHeaders(t *testing.T, e *echo.Echo, token string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	e.ServeHTTP(response, req)
	require.Equal(t, 503, response.Code)
	require.Equal(t, "1", response.Header().Get("Retry-After"))
	for range 20 {
		response = httptest.NewRecorder()
		e.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/me", nil))
	}
	require.Equal(t, 429, response.Code)
	require.Equal(t, "1", response.Header().Get("Retry-After"))
	require.Equal(t, "20", response.Header().Get("X-Ratelimit-Limit"))
	require.Equal(t, "no-store", response.Header().Get("Cache-Control"))
	require.NotEmpty(t, response.Header().Get("X-Request-ID"))
}
