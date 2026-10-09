//nolint:lll // Keep test-only provider wire data and rollback SQL readable without changing their representation.
package handler_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/database"
	"github.com/6sLOGAN78/flux/internal/errs"
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
	"github.com/jackc/pgx/v5/pgconn"
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
	return productRouterWithRandom(cfg, db, p, rand.Reader)
}

func productRouterWithRandom(cfg *config.Config, db *database.Database, p *signedProvider, random io.Reader) http.Handler {
	log := zerolog.New(zerolog.SyncWriter(&p.logs))
	srv := &server.Server{Config: cfg, Logger: &log, DB: db}
	cfg.Auth = config.AuthConfig{SecretKey: "test-only", Issuer: p.issuer, AuthorizedParties: []string{fixtureParty}}
	cfg.Links = config.LinksConfig{ManagedHost: "go.flux.test", BlockedHosts: []string{"blocked.example"}}
	clients := &clerk.ClientConfig{BackendConfig: clerk.BackendConfig{
		URL: clerk.String(p.server.URL), Key: clerk.String(cfg.Auth.SecretKey),
		HTTPClient: &http.Client{Timeout: 3 * time.Second}}}
	auth := service.NewAuthServiceWithClients(cfg.Auth, service.AuthClients{
		JWKS: jwks.NewClient(clients), Sessions: session.NewClient(clients), Users: user.NewClient(clients)})
	workspace := service.NewWorkspaceService(repository.NewWorkspaceRepository(db.Pool))
	return router.NewRouter(srv, &handler.Handlers{OpenAPI: handler.NewOpenAPIHandler(srv)}, &service.Services{Auth: auth,
		Identity:  service.NewIdentityService(repository.NewUserRepository(db.Pool), auth),
		Workspace: workspace, Links: service.NewLinkService(repository.NewLinkRepositoryWithRandom(db.Pool, random), workspace, cfg.Links)})
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
		workspaceAPI := httptest.NewServer(productRouter(db.Config, &database.Database{Pool: db.Pool}, p))
		defer workspaceAPI.Close()
		checkWorkspaceBootstrap(t, workspaceAPI.URL, db, p, token)
	})
	t.Run("workspace-restore", func(t *testing.T) {
		checkWorkspaceRestore(t, api.URL, db, p)
	})
	t.Run("custom-key", func(t *testing.T) {
		checkCustomKey(t, api.URL, db, p)
	})
	t.Run("link-library", func(t *testing.T) {
		checkLinkLibrary(t, api.URL, db, p)
	})

	t.Run("link-create", func(t *testing.T) {
		checkLinkCreate(t, api.URL, db, p)
	})
}

func checkCustomKey(t *testing.T, api string, db *fluxTesting.TestDB, p *signedProvider) {
	t.Helper()
	token := p.token(t, map[string]any{"sub": "user_custom", "sid": "sess_user_custom"})
	otherToken := p.token(t, map[string]any{"sub": "user_custom_other", "sid": "sess_user_custom_other"})
	tokens := []string{token, otherToken}
	paths := make([]string, 2)
	for i := range paths {
		code, result := workspaceRequest(t, api, tokens[i], "POST", "/workspaces", fmt.Sprintf("custom-workspace-%02d", i), `{"name":"Custom tenant"}`)
		require.Equal(t, 201, code)
		paths[i] = "/workspaces/" + result["workspace"].(map[string]any)["id"].(string) + "/links"
	}
	payload := `{"destination":"https://example.com/custom","customKey":"Launch_2026"}`
	code, first := workspaceRequest(t, api, token, "POST", paths[0], "custom-create-0001", payload)
	require.Equal(t, 201, code)
	require.Equal(t, "https://go.flux.test/launch_2026", first["link"].(map[string]any)["shortUrl"])
	code, replay := workspaceRequest(t, api, token, "POST", paths[0], "custom-create-0001", strings.ReplaceAll(payload, "Launch_2026", "launch_2026"))
	require.Equal(t, 201, code)
	require.Equal(t, first, replay)
	code, conflict := workspaceRequest(t, api, token, "POST", paths[0], "custom-create-0001", strings.ReplaceAll(payload, "Launch_2026", "changed-key"))
	require.Equal(t, 409, code)
	require.Equal(t, "REQUEST_REUSE_CONFLICT", conflict["code"])
	for i, key := range []string{"ab", strings.Repeat("a", 65), "-abc", "a/b", "a%2fb", "a b", "abcé", "Key", "api", "docs", "live", "ready", "static", "login", "register", "dashboard", "settings", "links", "admin"} {
		body, err := json.Marshal(map[string]string{"destination": "https://example.com", "customKey": key})
		require.NoError(t, err)
		code, _ = workspaceRequest(t, api, token, "POST", paths[0], fmt.Sprintf("custom-invalid-%02d", i), string(body))
		require.Equal(t, 400, code, key)
	}
	t.Run("global-concurrent-one-winner", func(t *testing.T) {
		responses := make(chan int, 2)
		var group sync.WaitGroup
		for i := range paths {
			group.Go(func() {
				status, response := workspaceRequest(t, api, tokens[i], "POST", paths[i], "custom-global-0001", `{"destination":"https://example.com","customKey":"GLOBAL-KEY"}`)
				if status == 409 {
					require.Equal(t, "KEY_UNAVAILABLE", response["code"])
					require.Equal(t, "This short key is unavailable. Choose another key or generate one.", response["message"])
				}
				responses <- status
			})
		}
		group.Wait()
		close(responses)
		statuses := []int{}
		for status := range responses {
			statuses = append(statuses, status)
		}
		require.ElementsMatch(t, []int{201, 409}, statuses)
		var count int
		require.NoError(t, db.Pool.QueryRow(context.Background(), "SELECT count(*) FROM links WHERE short_key='global-key'").Scan(&count))
		require.Equal(t, 1, count)
	})
	t.Run("same-request-concurrent-one-effect", func(t *testing.T) {
		var group sync.WaitGroup
		responses := make(chan map[string]any, 4)
		for range 4 {
			group.Go(func() {
				status, response := workspaceRequest(t, api, token, "POST", paths[0], "custom-concurrent-01", `{"destination":"https://example.com","customKey":"same-request"}`)
				require.Equal(t, 201, status)
				responses <- response
			})
		}
		group.Wait()
		close(responses)
		var committed map[string]any
		for response := range responses {
			if committed == nil {
				committed = response
			}
			require.Equal(t, committed, response)
		}
		var effects, ledger int
		require.NoError(t, db.Pool.QueryRow(context.Background(), "SELECT count(*) FROM links WHERE short_key='same-request'").Scan(&effects))
		require.NoError(t, db.Pool.QueryRow(context.Background(), "SELECT count(*) FROM mutation_requests WHERE request_key='custom-concurrent-01'").Scan(&ledger))
		require.Equal(t, 1, effects)
		require.Equal(t, 1, ledger)
	})
	t.Run("custom-effect-ledger-rollback", func(t *testing.T) {
		ctx := context.Background()
		_, err := db.Pool.Exec(ctx, `CREATE FUNCTION reject_custom_commit() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'PRIVATE-CUSTOM-COMMIT'; END $$; CREATE CONSTRAINT TRIGGER reject_custom_commit AFTER INSERT ON links DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION reject_custom_commit()`)
		require.NoError(t, err)
		defer func() {
			_, dropErr := db.Pool.Exec(ctx, "DROP TRIGGER reject_custom_commit ON links; DROP FUNCTION reject_custom_commit()")
			require.NoError(t, dropErr)
		}()
		failureCode, response := workspaceRequest(t, api, token, "POST", paths[0], "custom-rollback-01", `{"destination":"https://example.com","customKey":"rolled-back"}`)
		require.Equal(t, 503, failureCode)
		require.Equal(t, "Links temporarily unavailable", response["message"])
		var effects, ledger int
		require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM links WHERE short_key='rolled-back'").Scan(&effects))
		require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM mutation_requests WHERE request_key='custom-rollback-01'").Scan(&ledger))
		require.Zero(t, effects)
		require.Zero(t, ledger)
	})
	workspace := strings.TrimSuffix(strings.TrimPrefix(paths[0], "/workspaces/"), "/links")
	_, err := db.Pool.Exec(context.Background(), "DELETE FROM memberships WHERE workspace_id=$1", workspace)
	require.NoError(t, err)
	code, _ = workspaceRequest(t, api, token, "POST", paths[0], "custom-create-0001", payload)
	require.Equal(t, 404, code, "revoked actor cannot replay custom-key outcome")
}

func checkLinkCreate(t *testing.T, api string, db *fluxTesting.TestDB, p *signedProvider) {
	t.Helper()
	// Observe the process defaults used by an accidental destination fetch or
	// lookup. Only the actual loopback API and provider fixture traffic passes.
	// The probes surround real HTTP creation, its PostgreSQL commit, and corpus.
	resolver, transport := net.DefaultResolver, http.DefaultTransport
	var dns, destinationHTTP atomic.Uint64
	net.DefaultResolver = &net.Resolver{PreferGo: true, Dial: func(context.Context, string, string) (net.Conn, error) { //nolint:reassign // Sequential actual-creation egress observation, restored below.
		dns.Add(1)
		return nil, errors.New("destination DNS forbidden")
	}}
	http.DefaultTransport = destinationTransport{base: transport, requests: &destinationHTTP, //nolint:reassign // Sequential actual-creation egress observation, restored below.
		allowed: map[string]bool{strings.TrimPrefix(api, "http://"): true,
			strings.TrimPrefix(p.server.URL, "http://"): true}}
	t.Cleanup(func() {
		net.DefaultResolver, http.DefaultTransport = resolver, transport //nolint:reassign // Restore process defaults after the isolated registered subtest.
		require.Zero(t, dns.Load(), "creation must not resolve destination DNS")
		require.Zero(t, destinationHTTP.Load(), "creation must not fetch destinations")
	})
	probe, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, probeErr := net.DefaultResolver.LookupHost(probe, "destination-probe.example")
	require.Error(t, probeErr)
	require.Positive(t, dns.Load(), "DNS observation must be live")
	dns.Store(0)
	probeRequest, requestErr := http.NewRequestWithContext(probe, http.MethodGet, "https://destination-probe.example", nil)
	require.NoError(t, requestErr)
	_, probeErr = http.DefaultTransport.RoundTrip(probeRequest)
	require.Error(t, probeErr)
	require.Equal(t, uint64(1), destinationHTTP.Load(), "HTTP observation must be live")
	destinationHTTP.Store(0)
	ctx := context.Background()
	token := p.token(t, map[string]any{"sub": "user_links", "sid": "sess_user_links"})
	status, created := workspaceRequest(t, api, token, "POST", "/workspaces", "link-workspace-0001", `{"name":"Link tenant"}`)
	require.Equal(t, 201, status)
	workspace := created["workspace"].(map[string]any)["id"].(string)
	path := "/workspaces/" + workspace + "/links"
	payload := `{"destination":"https://example.com/path?q=ok","title":"<script>protected</script>"}`
	status, result := workspaceRequest(t, api, token, "POST", path, "link-create-00001", payload)
	require.Equal(t, 201, status, "generated link creation must be committed")
	link := result["link"].(map[string]any)
	require.Regexp(t, `^https://go.flux.test/[a-z2-7]{20}$`, link["shortUrl"])
	require.Equal(t, "1", link["version"])
	require.Equal(t, "active", link["lifecycle"])
	status, replay := workspaceRequest(t, api, token, "POST", path, "link-create-00001", payload)
	require.Equal(t, 201, status)
	require.Equal(t, result, replay)
	status, _ = workspaceRequest(t, api, token, "POST", path, "link-create-00001", `{"destination":"https://example.org"}`)
	require.Equal(t, 409, status)
	for i, destination := range []string{"https://BÜCHER.example/a%2Fb?q=a%2Bb#fragment", "http://8.8.8.8/path", "https://[2606:4700:4700::1111]/path"} {
		body, err := json.Marshal(map[string]string{"destination": destination})
		require.NoError(t, err)
		code, accepted := workspaceRequest(t, api, token, "POST", path, "link-public-0000"+string(rune('a'+i)), string(body))
		require.Equal(t, 201, code)
		if i == 0 {
			require.Equal(t, "https://xn--bcher-kva.example/a%2Fb?q=a%2Bb#fragment", accepted["link"].(map[string]any)["destination"])
		}
	}
	for i, destination := range []string{"//example.com", "https://user@example.com", "http://127.0.0.1", "http://169.254.169.254", "https://go.flux.test/a", "https://blocked.example", "https://example.com:99999", "https://example.com\\evil", "ftp://example.com", "http://[::ffff:127.0.0.1]", "http://192.88.99.1", "http://[fec0::1]", "http://[2001:20::1]", "http://[3fff::1]", "https://sub.blocked.example", "https://example.com/%0a"} {
		body, err := json.Marshal(map[string]string{"destination": destination})
		require.NoError(t, err)
		status, _ = workspaceRequest(t, api, token, "POST", path, "link-unsafe-0000"+string(rune('a'+i)), string(body))
		require.Equal(t, 400, status, destination)
	}
	status, detail := workspaceRequest(t, api, token, "GET", path+"/"+link["id"].(string), "", "")
	require.Equal(t, 200, status)
	require.Equal(t, result, detail)
	t.Run("injected-entropy-and-collision-policy", func(t *testing.T) {
		createWith := func(reader io.Reader, key string) (int, map[string]any) {
			cfg := *db.Config
			server := httptest.NewServer(productRouterWithRandom(&cfg, &database.Database{Pool: db.Pool}, p, reader))
			defer server.Close()
			guard := http.DefaultTransport.(destinationTransport)
			guard.allowed[strings.TrimPrefix(server.URL, "http://")] = true
			return workspaceRequest(t, server.URL, token, "POST", path, key, payload)
		}
		code, seeded := createWith(bytes.NewReader(make([]byte, 12)), "link-entropy-seed-01")
		require.Equal(t, 201, code)
		entropy := &countedEntropy{reader: bytes.NewReader(append(make([]byte, 12), bytes.Repeat([]byte{1}, 12)...))}
		code, collision := createWith(entropy, "link-entropy-retry-01")
		require.Equal(t, 201, code)
		require.Equal(t, 2, entropy.reads)
		require.NotEqual(t, seeded["link"].(map[string]any)["shortUrl"], collision["link"].(map[string]any)["shortUrl"])
		for _, item := range []struct {
			reader io.Reader
			key    string
			reads  int
		}{
			{strings.NewReader(""), "link-entropy-failure-01", 1},
			{bytes.NewReader(make([]byte, 11)), "link-entropy-short-01", 2},
			{bytes.NewReader(make([]byte, 60)), "link-entropy-exhaust-01", 5},
		} {
			var before, after, ledger int
			require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM links WHERE workspace_id=$1", workspace).Scan(&before))
			random := &countedEntropy{reader: item.reader}
			failureCode, failureResponse := createWith(random, item.key)
			require.Equal(t, 503, failureCode)
			require.Equal(t, item.reads, random.reads)
			require.Equal(t, "Links temporarily unavailable", failureResponse["message"])
			require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM links WHERE workspace_id=$1", workspace).Scan(&after))
			require.Equal(t, before, after)
			require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM mutation_requests WHERE request_key=$1", item.key).Scan(&ledger))
			require.Zero(t, ledger)
		}
	})
	for _, invalidBody := range []string{`null`, `{"destination":"https://example.com","title":null}`, `{"destination":"https://example.com","extra":true}`, `{"destination":"https://example.com"} {}`, `{"destination":"https://example.com","title":"` + strings.Repeat("x", 201) + `"}`} {
		status, _ = workspaceRequest(t, api, token, "POST", path, "link-invalid-0001", invalidBody)
		require.Equal(t, 400, status)
	}
	t.Run("concurrent-replay", func(t *testing.T) {
		var group sync.WaitGroup
		responses := make(chan map[string]any, 4)
		for range 4 {
			group.Go(func() {
				code, response := workspaceRequest(t, api, token, "POST", path, "link-concurrent-01", payload)
				require.Equal(t, 201, code)
				responses <- response
			})
		}
		group.Wait()
		close(responses)
		var first map[string]any
		for response := range responses {
			if first == nil {
				first = response
			}
			require.Equal(t, first, response)
		}
	})
	t.Run("atomic-commit-rollback", func(t *testing.T) {
		_, err := db.Pool.Exec(ctx, `CREATE FUNCTION reject_link_commit() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'PRIVATE-LINK-COMMIT'; END $$; CREATE CONSTRAINT TRIGGER reject_link_commit AFTER INSERT ON links DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION reject_link_commit()`)
		require.NoError(t, err)
		defer func() {
			_, dropErr := db.Pool.Exec(ctx, "DROP TRIGGER reject_link_commit ON links; DROP FUNCTION reject_link_commit()")
			require.NoError(t, dropErr)
		}()
		code, response := workspaceRequest(t, api, token, "POST", path, "link-rollback-0001", payload)
		require.Equal(t, 503, code)
		require.NotContains(t, response, "PRIVATE-LINK-COMMIT")
		var count int
		require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM mutation_requests WHERE request_key=$1", "link-rollback-0001").Scan(&count))
		require.Zero(t, count)
	})
	t.Run("global-key-collision-retries", func(t *testing.T) {
		_, updateErr := db.Pool.Exec(ctx, "UPDATE links SET lifecycle='deleted' WHERE workspace_id=$1 AND id=$2", workspace, link["id"])
		require.NoError(t, updateErr)
		_, err := db.Pool.Exec(ctx, `CREATE SEQUENCE link_collision_attempt; CREATE FUNCTION collide_link_key() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF nextval('link_collision_attempt') <= 2 THEN NEW.short_key := '`+strings.TrimPrefix(link["shortUrl"].(string), "https://go.flux.test/")+`'; END IF; RETURN NEW; END $$; CREATE TRIGGER collide_link_key BEFORE INSERT ON links FOR EACH ROW EXECUTE FUNCTION collide_link_key()`)
		require.NoError(t, err)
		defer func() {
			_, dropErr := db.Pool.Exec(ctx, "DROP TRIGGER collide_link_key ON links; DROP FUNCTION collide_link_key(); DROP SEQUENCE link_collision_attempt")
			require.NoError(t, dropErr)
		}()
		code, response := workspaceRequest(t, api, token, "POST", path, "link-collision-01", payload)
		require.Equal(t, 201, code)
		require.NotEqual(t, link["shortUrl"], response["link"].(map[string]any)["shortUrl"])
		var attempts int
		require.NoError(t, db.Pool.QueryRow(ctx, "SELECT last_value FROM link_collision_attempt").Scan(&attempts))
		require.Equal(t, 3, attempts)
	})
	t.Run("collision-exhaustion-bounded-to-five", func(t *testing.T) {
		_, err := db.Pool.Exec(ctx, `CREATE SEQUENCE link_exhaust_attempt; CREATE FUNCTION exhaust_link_key() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN PERFORM nextval('link_exhaust_attempt'); NEW.short_key := '`+strings.TrimPrefix(link["shortUrl"].(string), "https://go.flux.test/")+`'; RETURN NEW; END $$; CREATE TRIGGER exhaust_link_key BEFORE INSERT ON links FOR EACH ROW EXECUTE FUNCTION exhaust_link_key()`)
		require.NoError(t, err)
		defer func() {
			_, dropErr := db.Pool.Exec(ctx, "DROP TRIGGER exhaust_link_key ON links; DROP FUNCTION exhaust_link_key(); DROP SEQUENCE link_exhaust_attempt")
			require.NoError(t, dropErr)
		}()
		code, _ := workspaceRequest(t, api, token, "POST", path, "link-exhaust-0001", payload)
		require.Equal(t, 503, code)
		var attempts, ledger int
		require.NoError(t, db.Pool.QueryRow(ctx, "SELECT last_value FROM link_exhaust_attempt").Scan(&attempts))
		require.Equal(t, 5, attempts)
		require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM mutation_requests WHERE request_key='link-exhaust-0001'").Scan(&ledger))
		require.Zero(t, ledger)
	})
	t.Run("unrelated-unique-constraint-not-retried", func(t *testing.T) {
		_, err := db.Pool.Exec(ctx, `CREATE SEQUENCE link_other_attempt; CREATE FUNCTION reject_other_link_key() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN PERFORM nextval('link_other_attempt'); RAISE EXCEPTION 'PRIVATE-OTHER-CONSTRAINT' USING ERRCODE='23505', CONSTRAINT='unrelated_unique'; END $$; CREATE TRIGGER reject_other_link_key BEFORE INSERT ON links FOR EACH ROW EXECUTE FUNCTION reject_other_link_key()`)
		require.NoError(t, err)
		defer func() {
			_, dropErr := db.Pool.Exec(ctx, "DROP TRIGGER reject_other_link_key ON links; DROP FUNCTION reject_other_link_key(); DROP SEQUENCE link_other_attempt")
			require.NoError(t, dropErr)
		}()
		code, _ := workspaceRequest(t, api, token, "POST", path, "link-other-constraint-01", payload)
		require.Equal(t, 503, code)
		var attempts int
		require.NoError(t, db.Pool.QueryRow(ctx, "SELECT last_value FROM link_other_attempt").Scan(&attempts))
		require.Equal(t, 1, attempts)
	})
	t.Run("foreign-link-detail", func(t *testing.T) {
		code, other := workspaceRequest(t, api, token, "POST", "/workspaces", "link-other-ws-001", `{"name":"Other link tenant"}`)
		require.Equal(t, 201, code)
		otherID := other["workspace"].(map[string]any)["id"].(string)
		code, _ = workspaceRequest(t, api, token, "GET", "/workspaces/"+otherID+"/links/"+link["id"].(string), "", "")
		require.Equal(t, 404, code)
	})
	_, err := db.Pool.Exec(ctx, "UPDATE memberships SET role='viewer' WHERE workspace_id=$1", workspace)
	require.NoError(t, err)
	status, _ = workspaceRequest(t, api, token, "POST", path, "link-create-00001", payload)
	require.Equal(t, 403, status, "replay must require current write capability")
	status, _ = workspaceRequest(t, api, token, "GET", path+"/"+link["id"].(string), "", "")
	require.Equal(t, 200, status)
	_, err = db.Pool.Exec(ctx, "DELETE FROM memberships WHERE workspace_id=$1", workspace)
	require.NoError(t, err)
	status, _ = workspaceRequest(t, api, token, "GET", path+"/"+link["id"].(string), "", "")
	require.Equal(t, 404, status)
	status, _ = workspaceRequest(t, api, token, "POST", path, "link-create-00001", payload)
	require.Equal(t, 404, status)
	var creator string
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT creator_user_id::text FROM links WHERE workspace_id=$1 AND id=$2", workspace, link["id"]).Scan(&creator))
	require.Equal(t, link["creator"].(map[string]any)["id"], creator)
	otherToken := p.token(t, map[string]any{"sub": "user_link_reader", "sid": "sess_user_link_reader"})
	status, identity := workspaceRequest(t, api, otherToken, "GET", "/me", "", "")
	require.Equal(t, 200, status)
	reader := identity["user"].(map[string]any)["id"]
	_, err = db.Pool.Exec(ctx, "INSERT INTO memberships(workspace_id,user_id,role,created_by) VALUES($1,$2,'viewer',$2)", workspace, reader)
	require.NoError(t, err)
	status, detail = workspaceRequest(t, api, otherToken, "GET", path+"/"+link["id"].(string), "", "")
	require.Equal(t, 200, status)
	require.Equal(t, link["creator"], detail["link"].(map[string]any)["creator"], "authorized readers retain durable creator projection after removal")
	for _, marker := range []string{"PRIVATE-LINK-COMMIT", "PRIVATE-OTHER-CONSTRAINT", "<script>protected</script>", "https://example.com/path?q=ok"} {
		require.NotContains(t, p.logs.String(), marker)
	}
}

type destinationTransport struct {
	base     http.RoundTripper
	requests *atomic.Uint64
	allowed  map[string]bool
}

type countedEntropy struct {
	reader io.Reader
	reads  int
}

func (r *countedEntropy) Read(value []byte) (int, error) {
	r.reads++
	return r.reader.Read(value)
}

func (d destinationTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if !d.allowed[request.URL.Host] {
		d.requests.Add(1)
		return nil, errors.New("destination HTTP forbidden")
	}
	return d.base.RoundTrip(request)
}

func checkWorkspaceRestore(t *testing.T, api string, db *fluxTesting.TestDB, p *signedProvider) {
	t.Helper()
	ctx := context.Background()
	token := p.token(t, map[string]any{"sub": "user_restore", "sid": "sess_user_restore"})
	status, bootstrap := workspaceRequest(t, api, token, "GET", "/me", "", "")
	require.Equal(t, 200, status)
	require.Contains(t, bootstrap, "workspaces", "bootstrap must expose current memberships")
	require.Empty(t, bootstrap["workspaces"])
	require.Contains(t, bootstrap, "lastWorkspace")
	require.Nil(t, bootstrap["lastWorkspace"])
	status, created := workspaceRequest(t, api, token, "POST", "/workspaces", "restore-workspace-0001", `{"name":"Restore current"}`)
	require.Equal(t, 201, status)
	workspace := created["workspace"].(map[string]any)
	id := workspace["id"].(string)
	status, selected := workspaceRequest(t, api, token, "PUT", "/me/last-workspace", "", `{"workspaceId":"`+id+`"}`)
	require.Equal(t, 200, status)
	require.Equal(t, created, selected)
	status, bootstrap = workspaceRequest(t, api, token, "GET", "/me", "", "")
	require.Equal(t, 200, status)
	require.Equal(t, workspace, bootstrap["lastWorkspace"])
	require.Len(t, bootstrap["workspaces"], 1)
	t.Run("missing-resource-does-not-revoke-workspace", func(t *testing.T) {
		missing := uuid.NewString()
		requestStatus, absent := workspaceRequest(t, api, token, "GET", "/workspaces/"+missing, "", "")
		require.Equal(t, 404, requestStatus)
		require.NotContains(t, absent, "workspace")
		requestStatus, current := workspaceRequest(t, api, token, "GET", "/me", "", "")
		require.Equal(t, 200, requestStatus)
		require.Equal(t, workspace, current["lastWorkspace"])
		require.Len(t, current["workspaces"], 1)
		requestStatus, current = workspaceRequest(t, api, token, "GET", "/workspaces/"+id, "", "")
		require.Equal(t, 200, requestStatus)
		require.Equal(t, workspace, current["workspace"])
	})
	other := p.token(t, map[string]any{"sub": "user_restore_other", "sid": "sess_user_restore_other"})
	status, denied := workspaceRequest(t, api, other, "PUT", "/me/last-workspace", "", `{"workspaceId":"`+id+`"}`)
	require.Equal(t, 404, status)
	require.NotContains(t, denied, "workspace")
	for _, body := range []string{`{}`, `{"workspaceId":null}`, `{"workspaceId":"bad"}`, `{"workspaceId":"` + id + `","role":"owner"}`, `{"workspaceId":"` + id + `"} {}`} {
		status, _ = workspaceRequest(t, api, token, "PUT", "/me/last-workspace", "", body)
		require.Equal(t, 400, status)
	}
	var actor uuid.UUID
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT id FROM users WHERE subject=$1", "user_restore").Scan(&actor))
	t.Run("all-roles-can-select", func(t *testing.T) {
		for _, role := range []string{"owner", "admin", "member", "viewer"} {
			_, err := db.Pool.Exec(ctx, "UPDATE memberships SET role=$1 WHERE workspace_id=$2 AND user_id=$3", role, id, actor)
			require.NoError(t, err)
			requestStatus, result := workspaceRequest(t, api, token, "PUT", "/me/last-workspace", "", `{"workspaceId":"`+id+`"}`)
			require.Equal(t, 200, requestStatus)
			require.Equal(t, role, result["workspace"].(map[string]any)["role"])
		}
	})
	t.Run("removal-race-rechecks-before-commit", func(t *testing.T) {
		lock, err := db.Pool.Begin(ctx)
		require.NoError(t, err)
		defer func() { _ = lock.Rollback(context.Background()) }()
		_, err = repository.NewWorkspaceRepository(db.Pool).LockScope(ctx, lock, repository.Scope{WorkspaceID: uuid.MustParse(id), ActorID: actor}, true)
		require.NoError(t, err)
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, api+"/api/v1/me/last-workspace",
			strings.NewReader(`{"workspaceId":"`+id+`"}`))
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Origin", fixtureParty)
		req.Header.Set("Content-Type", "application/json")
		type outcome struct {
			response *http.Response
			err      error
		}
		finished := make(chan outcome, 1)
		go func() {
			response, requestErr := (&http.Client{Timeout: 5 * time.Second}).Do(req)
			finished <- outcome{response: response, err: requestErr}
		}()
		require.Eventually(t, func() bool {
			var waiting int
			return db.Pool.QueryRow(ctx, "SELECT count(*) FROM pg_stat_activity WHERE datname=current_database() AND wait_event_type='Lock' AND query LIKE 'SELECT id,name FROM workspaces%'").Scan(&waiting) == nil && waiting > 0
		}, time.Second, 10*time.Millisecond)
		_, err = lock.Exec(ctx, "DELETE FROM memberships WHERE workspace_id=$1 AND user_id=$2", id, actor)
		require.NoError(t, err)
		require.NoError(t, lock.Commit(ctx))
		result := <-finished
		require.NoError(t, result.err)
		defer result.response.Body.Close()
		require.Equal(t, 404, result.response.StatusCode)
		require.Equal(t, "no-store", result.response.Header.Get("Cache-Control"))
	})
	status, bootstrap = workspaceRequest(t, api, token, "GET", "/me", "", "")
	require.Equal(t, 200, status)
	require.Nil(t, bootstrap["lastWorkspace"])
	require.Empty(t, bootstrap["workspaces"])
	encoded, err := json.Marshal(bootstrap)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), id)
	require.NotContains(t, string(encoded), "Restore current")
}

func workspaceRequest(t *testing.T, api, token, method, path, key, body string) (int, map[string]any) {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), method, api+"/api/v1"+path, strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Origin", fixtureParty)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", key)
	// Separate test clients keep business assertions independent of burst limits.
	req.Header.Set("X-Forwarded-For", netip.AddrFrom16(uuid.New()).String())
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
	require.Equal(t, first["workspace"], summary["workspace"])
	require.Equal(t, "go.flux.test", summary["managedHost"])
	other := p.token(t, map[string]any{"sub": "user_other", "sid": "sess_other"})
	status, denied := workspaceRequest(t, api, other, "GET", "/workspaces/"+id, "", "")
	require.Equal(t, 404, status)
	require.NotContains(t, denied, "workspace")
	var group sync.WaitGroup
	for range 4 {
		group.Go(func() {
			requestStatus, result := workspaceRequest(t, api, token, "POST", "/workspaces", "workspace-concurrent-0001", `{"name":"Concurrent"}`)
			require.Equal(t, 201, requestStatus)
			require.Equal(t, "Concurrent", result["workspace"].(map[string]any)["name"])
		})
	}
	group.Wait()
	var rows int
	require.NoError(t, db.Pool.QueryRow(context.Background(), "SELECT count(*) FROM workspaces WHERE name=$1", "Concurrent").Scan(&rows))
	require.Equal(t, 1, rows)
	var owners int
	var durableActor uuid.UUID
	require.NoError(t, db.Pool.QueryRow(context.Background(), "SELECT user_id FROM memberships WHERE workspace_id=$1 AND role='owner'", id).Scan(&durableActor))
	require.NoError(t, db.Pool.QueryRow(context.Background(), "SELECT count(*) FROM memberships WHERE workspace_id=$1 AND role='owner'", id).Scan(&owners))
	require.Equal(t, 1, owners)
	var retention bool
	require.NoError(t, db.Pool.QueryRow(context.Background(), "SELECT retain_until >= created_at + interval '24 hours' AND octet_length(request_hash)=32 FROM workspace_bootstrap_requests WHERE workspace_id=$1", id).Scan(&retention))
	require.True(t, retention)
	t.Run("rollback-at-commit", func(t *testing.T) {
		_, err := db.Pool.Exec(context.Background(), `CREATE FUNCTION reject_workspace_commit() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'PRIVATE-WORKSPACE-COMMIT'; END $$;
CREATE CONSTRAINT TRIGGER reject_workspace_commit AFTER INSERT ON memberships DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION reject_workspace_commit()`)
		require.NoError(t, err)
		defer func() {
			_, dropErr := db.Pool.Exec(context.Background(), "DROP TRIGGER reject_workspace_commit ON memberships; DROP FUNCTION reject_workspace_commit()")
			require.NoError(t, dropErr)
		}()
		requestStatus, body := workspaceRequest(t, api, token, "POST", "/workspaces", "workspace-rollback-0001", `{"name":"Rolled back"}`)
		require.Equal(t, 503, requestStatus)
		require.NotContains(t, body, "workspace")
		require.NotContains(t, p.logs.String(), "PRIVATE-WORKSPACE-COMMIT")
		for _, statement := range []string{
			"SELECT count(*) FROM workspaces WHERE name='Rolled back'",
			"SELECT count(*) FROM workspace_bootstrap_requests WHERE request_key='workspace-rollback-0001'",
			"SELECT count(*) FROM memberships WHERE workspace_id NOT IN (SELECT id FROM workspaces)",
		} {
			require.NoError(t, db.Pool.QueryRow(context.Background(), statement).Scan(&rows))
			require.Zero(t, rows)
		}
	})
	t.Run("identity-key-scope-and-foreign-list", func(t *testing.T) {
		requestStatus, result := workspaceRequest(t, api, other, "POST", "/workspaces", "workspace-bootstrap-0001", `{"name":"Other team"}`)
		require.Equal(t, 201, requestStatus)
		require.NotEqual(t, id, result["workspace"].(map[string]any)["id"])
		requestStatus, result = workspaceRequest(t, api, other, "GET", "/workspaces", "", "")
		require.Equal(t, 200, requestStatus)
		require.Len(t, result["workspaces"], 1)
		require.Equal(t, "Other team", result["workspaces"].([]any)[0].(map[string]any)["name"])
	})
	t.Run("capabilities-and-composite-tenant-constraints", func(t *testing.T) {
		checkWorkspaceCapabilities(t, db, uuid.MustParse(id), durableActor)
	})
	t.Run("expired-cleanup-is-actor-scoped", func(t *testing.T) {
		_, err := db.Pool.Exec(context.Background(), "UPDATE workspace_bootstrap_requests SET "+
			"created_at=now()-interval '49 hours',retain_until=now()-interval '25 hours' WHERE request_key=$1", "workspace-bootstrap-0001")
		require.NoError(t, err)
		requestStatus, result := workspaceRequest(t, api, token, "POST", "/workspaces", "workspace-bootstrap-0001", `{"name":"After retention"}`)
		require.Equal(t, 201, requestStatus)
		require.NotEqual(t, id, result["workspace"].(map[string]any)["id"], "expired key may start a new operation")
		require.NoError(t, db.Pool.QueryRow(context.Background(), "SELECT count(*) FROM workspace_bootstrap_requests "+
			"WHERE actor_id<>$1 AND request_key=$2 AND retain_until<now()", durableActor, "workspace-bootstrap-0001").Scan(&rows))
		require.Equal(t, 1, rows, "cleanup cannot delete another actor's expired key")
		// Restore the original retry mapping for the following removal-denial corpus.
		_, err = db.Pool.Exec(context.Background(), "UPDATE workspace_bootstrap_requests SET workspace_id=$1,request_hash=$2 "+
			"WHERE actor_id=$3 AND request_key=$4", id, bootstrapHash("Growth 🚀"), durableActor, "workspace-bootstrap-0001")
		require.NoError(t, err)
	})
	t.Run("unverified-bootstrap-creates-no-workspace", func(t *testing.T) {
		provider := newSignedProvider(t)
		provider.profileMode = "unverified"
		unverifiedAPI := httptest.NewServer(productRouter(db.Config, &database.Database{Pool: db.Pool}, provider))
		defer unverifiedAPI.Close()
		requestStatus, body := workspaceRequest(t, unverifiedAPI.URL,
			provider.token(t, map[string]any{"sub": "user_bootstrap_unverified", "sid": "sess_user_bootstrap_unverified"}),
			"POST", "/workspaces", "workspace-unverified-01", `{"name":"Unverified"}`)
		require.Equal(t, 401, requestStatus)
		require.NotContains(t, body, "workspace")
		require.NoError(t, db.Pool.QueryRow(context.Background(), "SELECT count(*) FROM workspaces WHERE name=$1", "Unverified").Scan(&rows))
		require.Zero(t, rows)
	})
	t.Run("viewer-denied-write-and-removed-membership", func(t *testing.T) {
		var viewer uuid.UUID
		require.NoError(t, db.Pool.QueryRow(context.Background(), "SELECT id FROM users WHERE issuer=$1 AND subject=$2", fixtureIssuer, "user_other").Scan(&viewer))
		_, err := db.Pool.Exec(context.Background(), "INSERT INTO memberships(workspace_id,user_id,role,created_by) VALUES($1,$2,'viewer',$3)", id, viewer, durableActor)
		require.NoError(t, err)
		svc := service.NewWorkspaceService(repository.NewWorkspaceRepository(db.Pool))
		tx, err := db.Pool.Begin(context.Background())
		require.NoError(t, err)
		scope := repository.Scope{WorkspaceID: uuid.MustParse(id), ActorID: viewer}
		err = svc.RequireWrite(context.Background(), tx, scope)
		var denied *errs.HTTPError
		require.ErrorAs(t, err, &denied)
		require.Equal(t, 403, denied.Status)
		require.NoError(t, tx.Rollback(context.Background()))
		requestStatus, result := workspaceRequest(t, api, other, "GET", "/workspaces/"+id, "", "")
		require.Equal(t, 200, requestStatus)
		require.Equal(t, "viewer", result["workspace"].(map[string]any)["role"])
		_, err = db.Pool.Exec(context.Background(), "DELETE FROM memberships WHERE workspace_id=$1", id)
		require.NoError(t, err)
		requestStatus, _ = workspaceRequest(t, api, token, "GET", "/workspaces/"+id, "", "")
		require.Equal(t, 404, requestStatus)
		requestStatus, _ = workspaceRequest(t, api, token, "POST", "/workspaces", "workspace-bootstrap-0001", `{"name":"Growth 🚀"}`)
		require.Equal(t, 404, requestStatus, "replay must reauthorize current membership")
		requestStatus, result = workspaceRequest(t, api, token, "POST", "/workspaces", "workspace-bootstrap-0001", `{"name":"Changed"}`)
		require.Equal(t, 404, requestStatus, "removed actor cannot use a hash conflict to inspect the ledger")
		require.NotContains(t, result, "workspace")
	})
}

func bootstrapHash(name string) []byte {
	encoded, _ := json.Marshal(struct {
		Name string `json:"name"`
	}{name})
	hash := sha256.Sum256(encoded)
	return hash[:]
}

func checkWorkspaceCapabilities(t *testing.T, db *fluxTesting.TestDB, workspace, owner uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	var actor, foreignWorkspace uuid.UUID
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT id FROM users WHERE issuer=$1 AND subject=$2", fixtureIssuer, "user_other").Scan(&actor))
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT id FROM workspaces WHERE name=$1", "Other team").Scan(&foreignWorkspace))
	_, err := db.Pool.Exec(ctx, "INSERT INTO memberships(workspace_id,user_id,role,created_by) VALUES($1,$2,'member',$3)", workspace, actor, owner)
	require.NoError(t, err)
	svc := service.NewWorkspaceService(repository.NewWorkspaceRepository(db.Pool))
	_, err = db.Pool.Exec(ctx, "INSERT INTO audit_events(workspace_id,actor_id,operation) VALUES($1,$2,$3)", workspace, actor, "membership.provenance")
	require.NoError(t, err)
	_, err = db.Pool.Exec(ctx, "UPDATE memberships SET role=$1 WHERE workspace_id=$2 AND user_id=$3", "unknown", workspace, actor)
	var invalidRole *pgconn.PgError
	require.ErrorAs(t, err, &invalidRole)
	require.Equal(t, "23514", invalidRole.Code)
	for _, role := range []string{"owner", "admin", "member", "viewer"} {
		_, err = db.Pool.Exec(ctx, "UPDATE memberships SET role=$1 WHERE workspace_id=$2 AND user_id=$3", role, workspace, actor)
		require.NoError(t, err)
		for _, capability := range []service.Capability{service.CapabilityRead, service.CapabilityWrite, service.CapabilityTeam} {
			tx, beginErr := db.Pool.Begin(ctx)
			require.NoError(t, beginErr)
			permissionErr := svc.RequireCapability(ctx, tx, repository.Scope{WorkspaceID: workspace, ActorID: actor}, capability)
			if service.Allows(role, capability) {
				require.NoError(t, permissionErr)
			} else {
				var denied *errs.HTTPError
				require.ErrorAs(t, permissionErr, &denied)
				require.Equal(t, 403, denied.Status)
			}
			require.NoError(t, tx.Rollback(ctx))
		}
	}
	checkWorkspaceRevocationLock(t, db, svc, repository.Scope{WorkspaceID: workspace, ActorID: actor})
	// Composite keys reject a valid membership ID paired with a foreign tenant.
	tx, err := db.Pool.Begin(ctx)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback(context.Background()) }()
	_, err = tx.Exec(ctx, "CREATE TABLE tenant_constraint_probe (workspace_id uuid,id uuid, "+
		"FOREIGN KEY(workspace_id,id) REFERENCES memberships(workspace_id,id))")
	require.NoError(t, err)
	var membership uuid.UUID
	require.NoError(t, tx.QueryRow(ctx, "SELECT id FROM memberships WHERE workspace_id=$1 AND user_id=$2", workspace, actor).Scan(&membership))
	_, err = tx.Exec(ctx, "INSERT INTO tenant_constraint_probe VALUES($1,$2)", workspace, membership)
	require.NoError(t, err)
	_, err = tx.Exec(ctx, "INSERT INTO tenant_constraint_probe VALUES($1,$2)", foreignWorkspace, membership)
	var constraint *pgconn.PgError
	require.ErrorAs(t, err, &constraint)
	require.Equal(t, "23503", constraint.Code)
	require.NoError(t, tx.Rollback(ctx))
	_, err = db.Pool.Exec(ctx, "DELETE FROM memberships WHERE workspace_id=$1 AND user_id=$2", workspace, actor)
	require.NoError(t, err)
	// Historical creator identity is durable independently of removable membership.
	var creator uuid.UUID
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT created_by FROM workspaces WHERE id=$1", foreignWorkspace).Scan(&creator))
	require.Equal(t, actor, creator)
	var auditActor uuid.UUID
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT actor_id FROM audit_events WHERE workspace_id=$1 AND operation=$2", workspace, "membership.provenance").Scan(&auditActor))
	require.Equal(t, actor, auditActor, "membership removal cannot delete durable audit provenance")
}

func checkWorkspaceRevocationLock(t *testing.T, db *fluxTesting.TestDB, svc *service.WorkspaceService, scope repository.Scope) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := db.Pool.Exec(ctx, "UPDATE memberships SET role='member' WHERE workspace_id=$1 AND user_id=$2", scope.WorkspaceID, scope.ActorID)
	require.NoError(t, err)
	lock, err := db.Pool.Begin(ctx)
	require.NoError(t, err)
	defer func() { _ = lock.Rollback(context.Background()) }()
	_, err = repository.NewWorkspaceRepository(db.Pool).LockScope(ctx, lock, scope, true)
	require.NoError(t, err)
	finished := make(chan error, 1)
	go func() {
		tx, beginErr := db.Pool.Begin(ctx)
		if beginErr != nil {
			finished <- beginErr
			return
		}
		defer func() { _ = tx.Rollback(context.Background()) }()
		finished <- svc.RequireWrite(ctx, tx, scope)
	}()
	require.Eventually(t, func() bool {
		var waiting int
		return db.Pool.QueryRow(ctx, "SELECT count(*) FROM pg_stat_activity WHERE datname=current_database() AND wait_event_type='Lock'").Scan(&waiting) == nil && waiting > 0
	}, time.Second, 10*time.Millisecond)
	_, err = lock.Exec(ctx, "UPDATE memberships SET role='viewer' WHERE workspace_id=$1 AND user_id=$2", scope.WorkspaceID, scope.ActorID)
	require.NoError(t, err)
	require.NoError(t, lock.Commit(ctx))
	var denied *errs.HTTPError
	require.ErrorAs(t, <-finished, &denied)
	require.Equal(t, 403, denied.Status, "blocked write must re-read the committed revocation")
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

func checkLinkLibrary(t *testing.T, api string, db *fluxTesting.TestDB, p *signedProvider) {
	t.Helper()
	ctx := context.Background()
	token := p.token(t, nil)
	status, created := workspaceRequest(t, api, token, "POST", "/workspaces", "library-workspace-01", `{"name":"Library tenant"}`)
	require.Equal(t, 201, status)
	workspace := created["workspace"].(map[string]any)["id"].(string)
	path := "/workspaces/" + workspace + "/links"
	status, empty := workspaceRequest(t, api, token, "GET", path, "", "")
	require.Equal(t, 200, status, "collection must be registered on the production router")
	require.Empty(t, empty["items"])
	require.Nil(t, empty["nextCursor"])
	var actor string
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT id::text FROM users WHERE subject='user_fixture'").Scan(&actor))
	for i := range 105 {
		_, err := db.Pool.Exec(ctx, "INSERT INTO links (workspace_id,id,creator_user_id,managed_host,short_key,destination,title,created_at,lifecycle) VALUES($1,$2,$3,'go.flux.test',$4,'https://example.com/library',$5,'2026-01-01T12:00:00Z',$6)", workspace, fmt.Sprintf("00000000-0000-4000-8000-%012d", i+1), actor, fmt.Sprintf("library-key-%03d", i), fmt.Sprintf("Library %03d", i), map[bool]string{true: "deleted", false: "active"}[i == 104])
		require.NoError(t, err)
	}
	status, listed := workspaceRequest(t, api, token, "GET", path, "", "")
	require.Equal(t, 200, status)
	items := listed["items"].([]any)
	require.Len(t, items, 25)
	require.Len(t, listed, 2)
	require.NotEmpty(t, listed["nextCursor"], "an extra row must produce a signed continuation")
	for i, item := range items {
		link := item.(map[string]any)
		require.Equal(t, workspace, link["workspaceId"])
		require.Equal(t, fmt.Sprintf("Library %03d", 103-i), link["title"], "equal timestamps must sort by descending UUID")
		require.NotEqual(t, "deleted", link["lifecycle"])
	}
	for query, count := range map[string]int{"?limit=1": 1, "?limit=100": 100} {
		code, response := workspaceRequest(t, api, token, "GET", path+query, "", "")
		require.Equal(t, 200, code)
		require.Len(t, response["items"], count)
	}
	for _, query := range []string{"?limit=0", "?limit=101", "?limit=-1", "?limit=1.5", "?limit=abc", "?limit=1&limit=2", "?cursor=foreign", "?cursor=", "?workspaceId=" + uuid.NewString(), "?search=%27%20OR%201%3D1--", "?state=active"} {
		code, _ := workspaceRequest(t, api, token, "GET", path+query, "", "")
		require.Equal(t, 400, code, query)
	}
	seen := make(map[string]bool)
	page := listed
	for {
		for _, item := range page["items"].([]any) {
			id := item.(map[string]any)["id"].(string)
			require.False(t, seen[id], "equal timestamps must never duplicate rows")
			seen[id] = true
		}
		cursor, more := page["nextCursor"].(string)
		if !more {
			break
		}
		require.LessOrEqual(t, len(cursor), 2048)
		status, page = workspaceRequest(t, api, token, "GET", path+"?cursor="+url.QueryEscape(cursor), "", "")
		require.Equal(t, 200, status)
	}
	require.Len(t, seen, 104, "every nondeleted equal-timestamp row must appear")
	require.Len(t, page["items"], 4, "last page is authoritative exhaustion")
	cursor := listed["nextCursor"].(string)
	for _, invalid := range []string{"", "malformed", cursor[:len(cursor)-1] + "!", strings.Repeat("x", 2049)} {
		code, failure := workspaceRequest(t, api, token, "GET", path+"?cursor="+url.QueryEscape(invalid), "", "")
		require.Equal(t, 400, code)
		require.Equal(t, "CURSOR_INVALID", failure["code"])
		require.Equal(t, "This page is no longer available. Return to the first page.", failure["message"])
	}
	status, foreign := workspaceRequest(t, api, token, "POST", "/workspaces", "library-foreign-01", `{"name":"Foreign library"}`)
	require.Equal(t, 201, status)
	foreignID := foreign["workspace"].(map[string]any)["id"].(string)
	status, foreignCursor := workspaceRequest(t, api, token, "GET", "/workspaces/"+foreignID+"/links?cursor="+url.QueryEscape(cursor), "", "")
	require.Equal(t, 400, status)
	require.Equal(t, "CURSOR_INVALID", foreignCursor["code"])
	status, foreignList := workspaceRequest(t, api, token, "GET", "/workspaces/"+foreignID+"/links", "", "")
	require.Equal(t, 200, status)
	require.Empty(t, foreignList["items"], "foreign tenant rows never leak")
	status, _ = workspaceRequest(t, api, token, "GET", "/workspaces/"+foreignID+"/links/"+items[0].(map[string]any)["id"].(string), "", "")
	require.Equal(t, 404, status)
	_, err := db.Pool.Exec(ctx, "UPDATE memberships SET role='viewer' WHERE workspace_id=$1", workspace)
	require.NoError(t, err)
	status, _ = workspaceRequest(t, api, token, "GET", path, "", "")
	require.Equal(t, 200, status, "viewers may list")
	_, err = db.Pool.Exec(ctx, "DELETE FROM memberships WHERE workspace_id=$1", workspace)
	require.NoError(t, err)
	status, denied := workspaceRequest(t, api, token, "GET", path, "", "")
	require.Equal(t, 404, status, "fresh membership precedes listing")
	require.NotContains(t, fmt.Sprint(denied), "Library 103")
}
