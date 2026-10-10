//nolint:lll // Keep test-only provider wire data and rollback SQL readable without changing their representation.
package handler_test

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
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
	cfg.Links = config.LinksConfig{CursorKey: base64.StdEncoding.EncodeToString(make([]byte, 32)), ManagedHost: "go.flux.test", BlockedHosts: []string{"blocked.example"}}
	clients := &clerk.ClientConfig{BackendConfig: clerk.BackendConfig{
		URL: clerk.String(p.server.URL), Key: clerk.String(cfg.Auth.SecretKey),
		HTTPClient: &http.Client{Timeout: 3 * time.Second}}}
	auth := service.NewAuthServiceWithClients(cfg.Auth, service.AuthClients{
		JWKS: jwks.NewClient(clients), Sessions: session.NewClient(clients), Users: user.NewClient(clients)})
	workspace := service.NewWorkspaceService(repository.NewWorkspaceRepository(db.Pool))
	cfg.Invitations = config.InvitationConfig{ActiveKeyID: "fixture", EncryptionKeys: ` {"fixture":"` + base64.StdEncoding.EncodeToString(make([]byte, 32)) + `"}`, Sender: "invites@example.test", PublicOrigin: "https://app.flux.test"}
	team := service.NewTeamService(repository.NewTeamRepository(db.Pool), workspace)
	return router.NewRouter(srv, &handler.Handlers{OpenAPI: handler.NewOpenAPIHandler(srv)}, &service.Services{Auth: auth,
		Identity:    service.NewIdentityService(repository.NewUserRepository(db.Pool), auth),
		Team:        team,
		Invitations: service.NewInvitationService(repository.NewInvitationRepository(db.Pool), team, cfg.Invitations),
		Workspace:   workspace, Links: service.NewLinkService(repository.NewLinkRepositoryWithRandom(db.Pool, random), workspace, cfg.Links)})
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
	t.Run("link-search", func(t *testing.T) {
		checkLinkSearch(t, api.URL, db, p)
	})
	t.Run("team", func(t *testing.T) {
		checkTeam(t, api.URL, db, p)
	})
	t.Run("invitations", func(t *testing.T) {
		checkInvitations(t, api.URL, db, p)
	})
	t.Run("roles", func(t *testing.T) {
		checkTeamRoles(t, api.URL, db, p)
	})
	t.Run("removal", func(t *testing.T) {
		checkTeamRemoval(t, api.URL, db, p)
	})
	t.Run("link-library", func(t *testing.T) {
		checkLinkLibrary(t, api.URL, db, p)
	})

	t.Run("link-create", func(t *testing.T) {
		checkLinkCreate(t, api.URL, db, p)
	})
}

func checkInvitations(t *testing.T, api string, db *fluxTesting.TestDB, p *signedProvider) {
	t.Helper()
	ctx := context.Background()
	token := p.token(t, map[string]any{"sub": "user_invites", "sid": "sess_user_invites"})
	status, created := workspaceRequest(t, api, token, "POST", "/workspaces", "invite-workspace-01", `{"name":"Invitation tenant"}`)
	require.Equal(t, 201, status)
	w := created["workspace"].(map[string]any)["id"].(string)
	path := "/workspaces/" + w + "/invitations"
	status, queued := workspaceRequest(t, api, token, "POST", path, "invitation-create-01", `{"email":"  Future+Tag@Example.Test  ","role":"member"}`)
	require.Equal(t, 201, status, "a real authorized request must atomically queue an encrypted intent")
	if status != 201 {
		return
	}
	invite := queued["invitation"].(map[string]any)
	require.Equal(t, "future+tag@example.test", invite["email"])
	require.Equal(t, "Queued", invite["status"])
	var intents, ledgers, audits int
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM invitation_delivery_intents WHERE workspace_id=$1", w).Scan(&intents))
	require.Equal(t, 1, intents)
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM mutation_requests WHERE workspace_id=$1 AND operation='invitation.create'", w).Scan(&ledgers))
	require.Equal(t, 1, ledgers)
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM audit_events WHERE workspace_id=$1 AND operation='invitation.create'", w).Scan(&audits))
	require.Equal(t, 1, audits)
	status, replay := workspaceRequest(t, api, token, "POST", path, "invitation-create-01", `{"email":"future+tag@example.test","role":"member"}`)
	require.Equal(t, 201, status)
	require.Equal(t, queued, replay)
	var canonicalHash []byte
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT request_hash FROM mutation_requests WHERE workspace_id=$1 AND operation='invitation.create' AND request_key='invitation-create-01'", w).Scan(&canonicalHash))
	expectedHash := sha256.Sum256([]byte("future+tag@example.test\nmember"))
	require.Equal(t, expectedHash[:], canonicalHash)
	status, _ = workspaceRequest(t, api, token, "POST", path, "invitation-create-01", `{"email":"changed@example.test","role":"viewer"}`)
	require.Equal(t, 409, status)
	status, _ = workspaceRequest(t, api, token, "POST", path, "invitation-create-02", `{"email":"FUTURE+TAG@example.test","role":"member"}`)
	require.Equal(t, 409, status)
	status, listed := workspaceRequest(t, api, token, "GET", path, "", "")
	require.Equal(t, 200, status)
	require.Len(t, listed["items"], 1)
	t.Run("durable_status_projection", func(t *testing.T) {
		checkInvitationStatusProjection(t, api, db, token, w, invite["id"].(string))
	})
	for _, private := range []string{"token", "digest", "ciphertext", "keyId", "envelope"} {
		require.NotContains(t, fmt.Sprint(queued), private)
		require.NotContains(t, fmt.Sprint(listed), private)
	}
	var actor uuid.UUID
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT user_id FROM memberships WHERE workspace_id=$1", w).Scan(&actor))
	for _, role := range []string{"owner", "admin", "member", "viewer"} {
		for _, grant := range []string{"owner", "admin", "member", "viewer"} {
			_, roleErr := db.Pool.Exec(ctx, "UPDATE memberships SET role=$3 WHERE workspace_id=$1 AND user_id=$2", w, actor, role)
			require.NoError(t, roleErr)
			body := fmt.Sprintf(`{"email":"%s-%s@example.test","role":"%s"}`, role, grant, grant)
			status, _ = workspaceRequest(t, api, token, "POST", path, "invite-role-"+role+"-"+grant, body)
			expected := 403
			if grant == "owner" {
				expected = 400
			} else if role == "owner" || (role == "admin" && (grant == "member" || grant == "viewer")) {
				expected = 201
			}
			require.Equal(t, expected, status, role+" -> "+grant)
		}
	}
	_, demoteErr := db.Pool.Exec(ctx, "UPDATE memberships SET role='admin' WHERE workspace_id=$1 AND user_id=$2", w, actor)
	require.NoError(t, demoteErr)
	status, _ = workspaceRequest(t, api, token, "POST", path, "invite-role-owner-admin", `{"email":"changed@example.test","role":"member"}`)
	require.Equal(t, 403, status, "protected committed grant policy precedes a changed replay hash")
	_, restoreErr := db.Pool.Exec(ctx, "UPDATE memberships SET role=$3 WHERE workspace_id=$1 AND user_id=$2", w, actor, "owner")
	require.NoError(t, restoreErr)
	checkInvitationAtomicity(t, api, db, p, token, w, actor)
	_, err := db.Pool.Exec(ctx, "DELETE FROM memberships WHERE workspace_id=$1 AND user_id=$2", w, actor)
	require.NoError(t, err)
	status, _ = workspaceRequest(t, api, token, "POST", path, "invitation-create-01", `{"email":"changed@example.test","role":"viewer"}`)
	require.Equal(t, 404, status, "removed inviter is denied before replay hash")
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM invitations WHERE workspace_id=$1 AND inviter_id=$2", w, actor).Scan(&intents))
	require.Positive(t, intents, "durable inviter provenance survives membership deletion")
	require.NotContains(t, p.logs.String(), "future+tag@example.test")
}

// This verifies stored state projection only; no worker/provider delivery is claimed.
func checkInvitationStatusProjection(t *testing.T, api string, db *fluxTesting.TestDB, token, workspace, invite string) {
	t.Helper()
	ctx := context.Background()
	var originalCreated, originalExpires time.Time
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT created_at,expires_at FROM invitations WHERE workspace_id=$1 AND id=$2", workspace, invite).Scan(&originalCreated, &originalExpires))
	code, foreignCreated := workspaceRequest(t, api, token, "POST", "/workspaces", "invite-status-foreign-01", `{"name":"Foreign status tenant"}`)
	require.Equal(t, 201, code)
	foreign := foreignCreated["workspace"].(map[string]any)["id"].(string)
	_, err := db.Pool.Exec(ctx, "INSERT INTO invitations(workspace_id,id,inviter_id,email,role,token_digest) SELECT $3,id,inviter_id,email,role,token_digest FROM invitations WHERE workspace_id=$1 AND id=$2", workspace, invite, foreign)
	require.NoError(t, err)
	_, err = db.Pool.Exec(ctx, "INSERT INTO invitation_delivery_intents(workspace_id,id,invitation_id,key_id,ciphertext,state) SELECT $3,id,invitation_id,key_id,ciphertext,'failed' FROM invitation_delivery_intents WHERE workspace_id=$1 AND invitation_id=$2", workspace, invite, foreign)
	require.NoError(t, err)
	var originalCiphertext []byte
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT ciphertext FROM invitation_delivery_intents WHERE workspace_id=$1 AND invitation_id=$2", workspace, invite).Scan(&originalCiphertext))
	for _, test := range []struct{ state, delivery, status string }{
		{"pending", "queued", "Queued"}, {"pending", "leased", "Queued"},
		{"pending", "delivered", "Delivered"}, {"pending", "failed", "Failed"},
		{"accepted", "delivered", "Accepted"}, {"revoked", "delivered", "Revoked"},
		{"expired", "delivered", "Expired"},
	} {
		_, err = db.Pool.Exec(ctx, "UPDATE invitations SET state=$3 WHERE workspace_id=$1 AND id=$2", workspace, invite, test.state)
		require.NoError(t, err)
		_, err = db.Pool.Exec(ctx, "UPDATE invitation_delivery_intents SET state=$3,ciphertext=CASE WHEN $3='delivered' THEN NULL::bytea ELSE $4::bytea END,lease_until=CASE WHEN $3='leased' THEN now()+interval '1 minute' END,delivered_at=CASE WHEN $3='delivered' THEN now() END WHERE workspace_id=$1 AND invitation_id=$2", workspace, invite, test.delivery, originalCiphertext)
		require.NoError(t, err)
		statusCode, listed := workspaceRequest(t, api, token, "GET", "/workspaces/"+workspace+"/invitations", "", "")
		require.Equal(t, 200, statusCode)
		require.Len(t, listed["items"], 1)
		require.Equal(t, test.status, listed["items"].([]any)[0].(map[string]any)["status"], test.state+"/"+test.delivery)
	}
	_, err = db.Pool.Exec(ctx, "UPDATE invitations SET state='pending',created_at=now()-interval '8 days',expires_at=now()-interval '1 day' WHERE workspace_id=$1 AND id=$2", workspace, invite)
	require.NoError(t, err)
	code, elapsed := workspaceRequest(t, api, token, "GET", "/workspaces/"+workspace+"/invitations", "", "")
	require.Equal(t, 200, code)
	require.Equal(t, "Expired", elapsed["items"].([]any)[0].(map[string]any)["status"], "elapsed pending expiry precedes delivered projection")
	_, err = db.Pool.Exec(ctx, "UPDATE invitations SET state='pending',created_at=$3,expires_at=$4 WHERE workspace_id=$1 AND id=$2", workspace, invite, originalCreated, originalExpires)
	require.NoError(t, err)
	_, err = db.Pool.Exec(ctx, "UPDATE invitation_delivery_intents SET state='queued',ciphertext=$3,delivered_at=NULL WHERE workspace_id=$1 AND invitation_id=$2", workspace, invite, originalCiphertext)
	require.NoError(t, err)
}

func checkInvitationAtomicity(t *testing.T, api string, db *fluxTesting.TestDB, p *signedProvider, token, w string, actor uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	path := "/workspaces/" + w + "/invitations"
	for _, body := range []string{`null`, `{}`, `{"email":"a@example.test","role":null}`, `{"email":"a@example.test","role":"member","extra":"private"}`, `{"email":"a@example.test","role":"member"} {}`, `{"email":"bad","role":"member"}`, `{"email":"a@b","role":"member"}`} {
		status, _ := workspaceRequest(t, api, token, "POST", path, "invalid-invite-key-01", body)
		require.Equal(t, 400, status)
	}
	checkInvitationInputBounds(t, api, token, path, p)
	for _, query := range []string{"?unknown=x", "?after=bad", "?after=&after=bad"} {
		status, _ := workspaceRequest(t, api, token, "GET", path+query, "", "")
		require.Equal(t, 400, status)
	}
	status, _ := workspaceRequest(t, api, token, "POST", path+"?unknown=x", "invalid-invite-key-02", `{"email":"a@example.test","role":"member"}`)
	require.Equal(t, 400, status)
	foreign := uuid.NewString()
	status, _ = workspaceRequest(t, api, token, "GET", "/workspaces/"+foreign+"/invitations", "", "")
	require.Equal(t, 404, status)
	status, _ = workspaceRequest(t, api, token, "POST", "/workspaces/"+foreign+"/invitations", "invitation-create-01", `{"email":"a@example.test","role":"member"}`)
	require.Equal(t, 404, status)

	// Inspect real committed storage without printing any private material.
	var digest, ciphertext []byte
	var invite, delivery uuid.UUID
	var keyID string
	var created, expires time.Time
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT i.id,d.id,i.token_digest,d.ciphertext,d.key_id,i.created_at,i.expires_at FROM invitations i JOIN invitation_delivery_intents d ON d.workspace_id=i.workspace_id AND d.invitation_id=i.id WHERE i.workspace_id=$1 AND i.email='future+tag@example.test'", w).Scan(&invite, &delivery, &digest, &ciphertext, &keyID, &created, &expires))
	require.Equal(t, 7*24*time.Hour, expires.Sub(created))
	block, err := aes.NewCipher(make([]byte, 32))
	require.NoError(t, err)
	aead, err := cipher.NewGCMWithRandomNonce(block)
	require.NoError(t, err)
	aad := []byte("flux.invitation.v1\n" + w + "\n" + invite.String() + "\n" + delivery.String() + "\n" + keyID)
	plaintext, err := aead.Open(nil, nil, ciphertext, aad)
	require.NoError(t, err)
	defer clear(plaintext)
	var envelope map[string]string
	require.NoError(t, json.Unmarshal(plaintext, &envelope))
	raw, err := base64.RawURLEncoding.DecodeString(envelope["token"])
	require.NoError(t, err)
	require.Len(t, raw, 32, "stored token has 256 random bits")
	expected := sha256.Sum256([]byte(envelope["token"]))
	require.True(t, bytes.Equal(expected[:], digest), "only matching acceptance digest is stored")
	require.False(t, bytes.Contains(ciphertext, []byte(envelope["token"])))
	if strings.Contains(p.logs.String(), envelope["token"]) {
		t.Fatal("private token in logs")
	}
	_, err = aead.Open(nil, nil, ciphertext, []byte("foreign"))
	require.Error(t, err)
	var snapshots string
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT string_agg(result::text,'') FROM mutation_requests WHERE workspace_id=$1", w).Scan(&snapshots))
	if strings.Contains(snapshots, envelope["token"]) {
		t.Fatal("private token in replay snapshot")
	}
	var stored string
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT row_to_json(i)::text || row_to_json(d)::text FROM invitations i JOIN invitation_delivery_intents d ON d.workspace_id=i.workspace_id AND d.invitation_id=i.id WHERE i.workspace_id=$1 AND i.id=$2", w, invite).Scan(&stored))
	if strings.Contains(stored, envelope["token"]) {
		t.Fatal("plaintext credential stored in relational row")
	}
	for _, label := range []string{"token", "ciphertext", "keyId", "digest"} {
		require.NotContains(t, snapshots, label)
	}

	// Workspace lock serializes simultaneous normalized-email creation.
	var wait sync.WaitGroup
	codes := make(chan int, 2)
	for i := range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			codes <- concurrentInvitationRequest(api, token, path, fmt.Sprintf("invite-concurrent-%02d", i))
		}()
	}
	wait.Wait()
	close(codes)
	var successes, duplicates int
	for code := range codes {
		switch code {
		case 201:
			successes++
		case 409:
			duplicates++
		}
	}
	require.Equal(t, 1, successes)
	require.Equal(t, 1, duplicates)
	// Old elapsed pending invite releases uniqueness inside the new transaction.
	_, err = db.Pool.Exec(ctx, "UPDATE invitations SET created_at=now()-interval '8 days',expires_at=now()-interval '1 day' WHERE workspace_id=$1 AND email='race@example.test'", w)
	require.NoError(t, err)
	checkInvitationExpiryRace(t, api, token, path)
	var count int
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM invitations WHERE workspace_id=$1 AND email='race@example.test' AND state='expired'", w).Scan(&count))
	require.Equal(t, 1, count)

	// A failed ledger insert rolls back all preceding invitation/intent/audit writes.
	_, err = db.Pool.Exec(ctx, "ALTER TABLE mutation_requests ADD CONSTRAINT deny_invite_test CHECK (request_key<>'invite-rollback-key-01')")
	require.NoError(t, err)
	status, _ = workspaceRequest(t, api, token, "POST", path, "invite-rollback-key-01", `{"email":"rollback@example.test","role":"member"}`)
	require.Equal(t, 503, status)
	_, err = db.Pool.Exec(ctx, "ALTER TABLE mutation_requests DROP CONSTRAINT deny_invite_test")
	require.NoError(t, err)
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM invitations WHERE workspace_id=$1 AND email='rollback@example.test'", w).Scan(&count))
	require.Zero(t, count)
	_, err = db.Pool.Exec(ctx, `CREATE FUNCTION deny_invitation_commit_test() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.email='commit-fail@example.test' THEN RAISE EXCEPTION 'synthetic commit failure'; END IF; RETURN NEW; END $$;
CREATE CONSTRAINT TRIGGER deny_invitation_commit_test AFTER INSERT ON invitations DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION deny_invitation_commit_test()`)
	require.NoError(t, err)
	status, _ = workspaceRequest(t, api, token, "POST", path, "invite-commit-failure-01", `{"email":"commit-fail@example.test","role":"member"}`)
	require.Equal(t, 503, status)
	_, err = db.Pool.Exec(ctx, "DROP TRIGGER deny_invitation_commit_test ON invitations; DROP FUNCTION deny_invitation_commit_test()")
	require.NoError(t, err)
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM invitations WHERE workspace_id=$1 AND email='commit-fail@example.test'", w).Scan(&count))
	require.Zero(t, count)
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM mutation_requests WHERE workspace_id=$1 AND request_key='invite-commit-failure-01'", w).Scan(&count))
	require.Zero(t, count)
	var retained bool
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT bool_and(retain_until>=created_at+interval '24 hours') FROM mutation_requests WHERE workspace_id=$1 AND operation='invitation.create'", w).Scan(&retained))
	require.True(t, retained)
	_, err = db.Pool.Exec(ctx, "UPDATE audit_events SET operation='tampered' WHERE workspace_id=$1 AND operation='invitation.create'", w)
	require.Error(t, err)

	// Exhaust every bounded list page; continuations never authorize another tenant.
	_, err = db.Pool.Exec(ctx, "INSERT INTO invitations(workspace_id,id,inviter_id,email,role,token_digest) SELECT $1,gen_random_uuid(),$2,'invite-page-'||n||'@example.test','member',decode(repeat('00',32),'hex') FROM generate_series(1,30) n", w, actor)
	require.NoError(t, err)
	seen := make(map[string]bool)
	queryPath := path
	for {
		code, page := workspaceRequest(t, api, token, "GET", queryPath, "", "")
		require.Equal(t, 200, code)
		items := page["items"].([]any)
		require.LessOrEqual(t, len(items), 25)
		encoded, encodeErr := json.Marshal(page)
		require.NoError(t, encodeErr)
		require.Less(t, len(encoded), 65536)
		for _, rawItem := range items {
			item := rawItem.(map[string]any)
			id := item["id"].(string)
			require.False(t, seen[id])
			seen[id] = true
			require.Equal(t, w, item["workspaceId"])
		}
		if page["nextAfter"] == nil {
			break
		}
		queryPath = path + "?after=" + page["nextAfter"].(string)
	}
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM invitations WHERE workspace_id=$1", w).Scan(&count))
	require.Len(t, seen, count)
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM invitation_delivery_intents d LEFT JOIN invitations i ON i.workspace_id=d.workspace_id AND i.id=d.invitation_id WHERE d.workspace_id=$1 AND i.id IS NULL", w).Scan(&count))
	require.Zero(t, count)

	// Cancellation while waiting for the workspace lock produces no effect.
	lock, err := db.Pool.Begin(ctx)
	require.NoError(t, err)
	_, err = lock.Exec(ctx, "SELECT id FROM workspaces WHERE id=$1 FOR UPDATE", w)
	require.NoError(t, err)
	cancelCtx, cancel := context.WithCancel(ctx)
	request, err := http.NewRequestWithContext(cancelCtx, http.MethodPost, api+"/api/v1"+path, strings.NewReader(`{"email":"cancelled@example.test","role":"member"}`))
	require.NoError(t, err)
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Origin", fixtureParty)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "invite-cancelled-key-01")
	done := make(chan error, 1)
	go func() {
		response, requestErr := http.DefaultClient.Do(request)
		if response != nil {
			response.Body.Close()
		}
		done <- requestErr
	}()
	require.Eventually(t, func() bool {
		var waiting int
		return db.Pool.QueryRow(ctx, "SELECT count(*) FROM pg_stat_activity WHERE datname=current_database() AND wait_event_type='Lock'").Scan(&waiting) == nil && waiting > 0
	}, 2*time.Second, 20*time.Millisecond)
	cancel()
	require.Error(t, <-done)
	require.NoError(t, lock.Rollback(ctx))
	require.Eventually(t, func() bool {
		var idle int
		return db.Pool.QueryRow(ctx, "SELECT count(*) FROM pg_stat_activity WHERE datname=current_database() AND state='idle in transaction'").Scan(&idle) == nil && idle == 0
	}, 2*time.Second, 20*time.Millisecond)
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM invitations WHERE workspace_id=$1 AND email='cancelled@example.test'", w).Scan(&count))
	require.Zero(t, count)
	// Durable users, not removable memberships, retain inviter identity.
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM invitations WHERE workspace_id=$1 AND inviter_id=$2", w, actor).Scan(&count))
	require.Positive(t, count)
}

func checkInvitationInputBounds(t *testing.T, api, token, path string, p *signedProvider) {
	t.Helper()
	for _, email := range []string{
		"name <private-canary@example.test>", "private-canary@example.test\r\nInjected: value",
		strings.Repeat("a", 242) + "@example.test", strings.Repeat(" ", 65536) + "private-canary@example.test",
	} {
		body, err := json.Marshal(map[string]string{"email": email, "role": "member"})
		require.NoError(t, err)
		code, _ := workspaceRequest(t, api, token, "POST", path, "invalid-invite-bound-01", string(body))
		require.Contains(t, []int{400, 413}, code)
	}
	require.NotContains(t, p.logs.String(), "private-canary")
}

func concurrentInvitationRequest(api, token, path, key string) int {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, api+"/api/v1"+path,
		strings.NewReader(`{"email":"race@example.test","role":"member"}`))
	if err != nil {
		return 0
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Origin", fixtureParty)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", key)
	request.Header.Set("X-Forwarded-For", netip.AddrFrom16(uuid.New()).String())
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return 0
	}
	defer response.Body.Close()
	_, err = io.Copy(io.Discard, response.Body)
	if err != nil {
		return 0
	}
	return response.StatusCode
}

func checkTeamRemoval(t *testing.T, api string, db *fluxTesting.TestDB, p *signedProvider) {
	t.Helper()
	ctx := context.Background()
	token := p.token(t, map[string]any{"sub": "user_removal", "sid": "sess_user_removal", "org_role": "org:owner"})
	other := p.token(t, map[string]any{"sub": "user_removal_other", "sid": "sess_user_removal_other"})
	status, created := workspaceRequest(t, api, token, "POST", "/workspaces", "removal-workspace-01", `{"name":"Removal tenant"}`)
	require.Equal(t, 201, status)
	w := created["workspace"].(map[string]any)["id"].(string)
	status, foreign := workspaceRequest(t, api, other, "POST", "/workspaces", "removal-workspace-02", `{"name":"Foreign removal tenant"}`)
	require.Equal(t, 201, status)
	f := foreign["workspace"].(map[string]any)["id"].(string)
	var actor, target string
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT user_id::text FROM memberships WHERE workspace_id=$1", w).Scan(&actor))
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT user_id::text FROM memberships WHERE workspace_id=$1", f).Scan(&target))
	path := "/workspaces/" + w + "/members/" + target
	self := "/workspaces/" + w + "/members/" + actor
	t.Run("foreign-target", func(t *testing.T) {
		code, _ := workspaceRequest(t, api, token, "DELETE", path, "removal-foreign-0001", `{}`)
		require.Equal(t, 404, code)
	})
	checkRemovalMatrix(t, api, db, token, w, actor, target, path)
	_, err := db.Pool.Exec(ctx, "UPDATE memberships SET role='owner' WHERE workspace_id=$1 AND user_id=$2", w, actor)
	require.NoError(t, err)
	_, err = db.Pool.Exec(ctx, "INSERT INTO memberships(workspace_id,user_id,role,created_by) VALUES($1,$2,'member',$3) ON CONFLICT(workspace_id,user_id) DO UPDATE SET role='member'", w, target, actor)
	require.NoError(t, err)
	status, denied := workspaceRequest(t, api, token, "DELETE", self, "removal-final-00001", `{}`)
	require.Equal(t, 409, status)
	require.Equal(t, "OWNER_REQUIRED", denied["code"])
	for _, body := range []string{`{"role":"owner"}`, `null`, `{} {}`} {
		status, _ = workspaceRequest(t, api, token, "DELETE", path, "removal-invalid-001", body)
		require.Equal(t, 400, status)
	}
	status, _ = workspaceRequest(t, api, token, "DELETE", path+"?role=owner", "removal-invalid-001", `{}`)
	require.Equal(t, 400, status)
	status, _ = workspaceRequest(t, api, token, "PATCH", path, "removal-replay-0001", `{"role":"member"}`)
	require.Equal(t, 200, status, "the same key belongs to a separate member.role operation")
	// Both lifecycle rows and their original replay snapshots belong to a durable user.
	var links []map[string]any
	for _, key := range []string{"remove-active", "remove-deleted"} {
		linkStatus, response := workspaceRequest(t, api, other, "POST", "/workspaces/"+w+"/links", "removal-link-"+key,
			`{"destination":"https://example.com/retained","customKey":"`+key+`"}`)
		require.Equal(t, 201, linkStatus)
		links = append(links, response["link"].(map[string]any))
	}
	_, err = db.Pool.Exec(ctx, "UPDATE links SET lifecycle='deleted' WHERE workspace_id=$1 AND id=$2", w, links[1]["id"])
	require.NoError(t, err)
	status, before := workspaceRequest(t, api, token, "GET", "/workspaces/"+w+"/links/"+links[1]["id"].(string), "", "")
	require.Equal(t, 200, status)
	links[1] = before["link"].(map[string]any)
	var membership string
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT id::text FROM memberships WHERE workspace_id=$1 AND user_id=$2", w, target).Scan(&membership))
	checkRemovalRollback(t, api, db, token, w, target, path)
	status, first := workspaceRequest(t, api, token, "DELETE", path, "removal-replay-0001", `{}`)
	require.Equal(t, 200, status)
	require.Equal(t, target, first["removedUserId"])
	require.Equal(t, w, first["workspaceId"])
	require.Equal(t, false, first["selfRemoved"])
	status, replay := workspaceRequest(t, api, token, "DELETE", path, "removal-replay-0001", `{}`)
	require.Equal(t, 200, status, "authorized replay works after the membership is physically absent")
	require.Equal(t, first, replay)
	status, _ = workspaceRequest(t, api, token, "DELETE", self, "removal-replay-0001", `{}`)
	require.Equal(t, 409, status, "changed target conflicts without removing another member")
	status, _ = workspaceRequest(t, api, token, "DELETE", path, "removal-missing-0001", `{}`)
	require.Equal(t, 404, status, "a new request cannot recover a removed target")
	var count int
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM links WHERE workspace_id=$1 AND creator_user_id=$2", w, target).Scan(&count))
	require.Equal(t, 2, count)
	for _, link := range links {
		linkStatus, response := workspaceRequest(t, api, token, "GET", "/workspaces/"+w+"/links/"+link["id"].(string), "", "")
		require.Equal(t, 200, linkStatus)
		require.Equal(t, link, response["link"], "UUID, canonical URL, lifecycle, creator and safe projection survive")
		status, _ = workspaceRequest(t, api, other, "GET", "/workspaces/"+w+"/links/"+link["id"].(string), "", "")
		require.Equal(t, 404, status)
	}
	for _, endpoint := range []string{"", "/members", "/links"} {
		status, _ = workspaceRequest(t, api, other, "GET", "/workspaces/"+w+endpoint, "", "")
		require.Equal(t, 404, status)
	}
	status, _ = workspaceRequest(t, api, other, "POST", "/workspaces/"+w+"/links", "removal-link-remove-active", `{"destination":"https://example.com/retained","customKey":"remove-active"}`)
	require.Equal(t, 404, status, "removed creator cannot replay its old create")
	status, _ = workspaceRequest(t, api, other, "PATCH", self, "removal-oldrole-001", `{"role":"owner"}`)
	require.Equal(t, 404, status)
	status, _ = workspaceRequest(t, api, other, "DELETE", self, "removal-replay-0001", `{}`)
	require.Equal(t, 404, status)
	status, _ = workspaceRequest(t, api, other, "PUT", "/me/last-workspace", "", `{"workspaceId":"`+w+`"}`)
	require.Equal(t, 404, status)
	status, _ = workspaceRequest(t, api, token, "POST", "/workspaces/"+w+"/links", "removal-reserved-001", `{"destination":"https://example.com/retained","customKey":"remove-deleted"}`)
	require.Equal(t, 409, status, "deleted key reservation remains")
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM audit_events WHERE workspace_id=$1 AND actor_id=$2 AND operation='member.remove' AND target_membership_id=$3 AND created_at IS NOT NULL", w, actor, membership).Scan(&count))
	require.Equal(t, 1, count)
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM mutation_requests WHERE workspace_id=$1 AND actor_id=$2 AND operation='member.remove' AND request_key='removal-replay-0001' AND retain_until>=created_at+interval '24 hours'", w, actor).Scan(&count))
	require.Equal(t, 1, count)
	// An admin cannot replay a snapshot of a former owner, even with a matching hash.
	_, err = db.Pool.Exec(ctx, "UPDATE memberships SET role='admin' WHERE workspace_id=$1 AND user_id=$2", w, actor)
	require.NoError(t, err)
	status, _ = workspaceRequest(t, api, token, "DELETE", path, "removal-replay-0001", `{}`)
	require.Equal(t, 200, status, "current admin may replay removal of a former member")
	_, err = db.Pool.Exec(ctx, "UPDATE memberships SET role='member' WHERE workspace_id=$1 AND user_id=$2", w, actor)
	require.NoError(t, err)
	status, _ = workspaceRequest(t, api, token, "DELETE", self, "removal-replay-0001", `{}`)
	require.Equal(t, 403, status, "fresh actor policy precedes private changed-hash comparison")
	_, err = db.Pool.Exec(ctx, "UPDATE memberships SET role='owner' WHERE workspace_id=$1 AND user_id=$2", w, actor)
	require.NoError(t, err)
	_, err = db.Pool.Exec(ctx, "INSERT INTO memberships(workspace_id,user_id,role,created_by) VALUES($1,$2,'owner',$3)", w, target, actor)
	require.NoError(t, err)
	checkRemovalRace(t, api, db, token, other, w, actor, target)
	checkRemovalSnapshotPolicy(t, api, db, token, w, actor, target, path)
	// Still authorized owner self-removal commits, but that former actor cannot replay.
	_, err = db.Pool.Exec(ctx, "INSERT INTO memberships(workspace_id,user_id,role,created_by) VALUES($1,$2,'owner',$2) ON CONFLICT(workspace_id,user_id) DO UPDATE SET role='owner'", w, actor)
	require.NoError(t, err)
	_, err = db.Pool.Exec(ctx, "INSERT INTO memberships(workspace_id,user_id,role,created_by) VALUES($1,$2,'owner',$2) ON CONFLICT(workspace_id,user_id) DO UPDATE SET role='owner'", w, target)
	require.NoError(t, err)
	status, response := workspaceRequest(t, api, token, "DELETE", self, "removal-self-000001", `{}`)
	require.Equal(t, 200, status)
	require.Equal(t, true, response["selfRemoved"])
	status, _ = workspaceRequest(t, api, token, "DELETE", self, "removal-self-000001", `{}`)
	require.Equal(t, 404, status)
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM audit_events WHERE workspace_id=$1 AND actor_id=$2 AND operation='member.remove'", w, actor).Scan(&count))
	require.Positive(t, count, "durable actor audit survives self-removal")
	_, err = db.Pool.Exec(ctx, "UPDATE audit_events SET operation='changed' WHERE workspace_id=$1", w)
	require.Error(t, err)
	_, err = db.Pool.Exec(ctx, "DELETE FROM audit_events WHERE workspace_id=$1", w)
	require.Error(t, err)
	require.NotContains(t, p.logs.String(), "PRIVATE-REMOVAL-COMMIT")
}

func checkRemovalSnapshotPolicy(t *testing.T, api string, db *fluxTesting.TestDB, token, w, actor, target, path string) {
	t.Helper()
	ctx := context.Background()
	for _, user := range []string{actor, target} {
		_, err := db.Pool.Exec(ctx, "INSERT INTO memberships(workspace_id,user_id,role,created_by) VALUES($1,$2,'owner',$2) ON CONFLICT(workspace_id,user_id) DO UPDATE SET role='owner'", w, user)
		require.NoError(t, err)
	}
	status, _ := workspaceRequest(t, api, token, "DELETE", path, "removal-owner-00001", `{}`)
	require.Equal(t, 200, status)
	_, err := db.Pool.Exec(ctx, "UPDATE memberships SET role='admin' WHERE workspace_id=$1 AND user_id=$2", w, actor)
	require.NoError(t, err)
	status, _ = workspaceRequest(t, api, token, "DELETE", path, "removal-owner-00001", `{}`)
	require.Equal(t, 403, status, "current admin cannot replay protected former-owner snapshot")
	status, _ = workspaceRequest(t, api, token, "DELETE", "/workspaces/"+w+"/members/"+uuid.NewString(), "removal-owner-00001", `{}`)
	require.Equal(t, 403, status, "snapshot policy precedes changed-hash disclosure for missing target")
	_, err = db.Pool.Exec(ctx, "UPDATE memberships SET role='owner' WHERE workspace_id=$1 AND user_id=$2", w, actor)
	require.NoError(t, err)
	lock, err := db.Pool.Begin(ctx)
	require.NoError(t, err)
	defer lock.Rollback(ctx)
	_, err = lock.Exec(ctx, "SELECT id FROM workspaces WHERE id=$1 FOR UPDATE", w)
	require.NoError(t, err)
	results := make(chan roleRequestResult, 1)
	go func() { results <- removalStatus(api, token, path, "removal-owner-00001") }()
	require.Eventually(t, func() bool {
		var waiting int
		return db.Pool.QueryRow(ctx, "SELECT count(*) FROM pg_stat_activity WHERE datname=current_database() AND wait_event_type='Lock'").Scan(&waiting) == nil && waiting >= 1
	}, 2*time.Second, 10*time.Millisecond)
	_, err = lock.Exec(ctx, "DELETE FROM memberships WHERE workspace_id=$1 AND user_id=$2", w, actor)
	require.NoError(t, err)
	require.NoError(t, lock.Commit(ctx))
	result := <-results
	require.NoError(t, result.err)
	require.Equal(t, 404, result.code, "pending replay observes committed actor removal before reading ledger")
}

func checkRemovalMatrix(t *testing.T, api string, db *fluxTesting.TestDB, token, w, actor, target, path string) {
	t.Helper()
	for _, actorRole := range []string{"owner", "admin", "member", "viewer"} {
		for _, targetRole := range []string{"owner", "admin", "member", "viewer"} {
			t.Run(actorRole+"-removes-"+targetRole, func(t *testing.T) {
				ctx := context.Background()
				_, err := db.Pool.Exec(ctx, "UPDATE memberships SET role=$3 WHERE workspace_id=$1 AND user_id=$2", w, actor, actorRole)
				require.NoError(t, err)
				_, err = db.Pool.Exec(ctx, "INSERT INTO memberships(workspace_id,user_id,role,created_by) VALUES($1,$2,$3,$4) ON CONFLICT(workspace_id,user_id) DO UPDATE SET role=EXCLUDED.role", w, target, targetRole, actor)
				require.NoError(t, err)
				status, _ := workspaceRequest(t, api, token, "DELETE", path, "removal-matrix-"+uuid.NewString(), `{}`)
				allowed := actorRole == "owner" || actorRole == "admin" && (targetRole == "member" || targetRole == "viewer")
				if allowed {
					require.Equal(t, 200, status)
				} else {
					require.Equal(t, 403, status)
				}
				time.Sleep(70 * time.Millisecond)
			})
		}
	}
}

func checkRemovalRollback(t *testing.T, api string, db *fluxTesting.TestDB, token, w, target, path string) {
	t.Helper()
	ctx := context.Background()
	var auditsBefore int
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM audit_events WHERE workspace_id=$1", w).Scan(&auditsBefore))
	_, err := db.Pool.Exec(ctx, `CREATE FUNCTION reject_removal_commit() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'PRIVATE-REMOVAL-COMMIT'; END $$;
CREATE CONSTRAINT TRIGGER reject_removal_commit AFTER DELETE ON memberships DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION reject_removal_commit()`)
	require.NoError(t, err)
	status, _ := workspaceRequest(t, api, token, "DELETE", path, "removal-rollback-01", `{}`)
	require.Equal(t, 503, status)
	_, err = db.Pool.Exec(ctx, "DROP TRIGGER reject_removal_commit ON memberships; DROP FUNCTION reject_removal_commit()")
	require.NoError(t, err)
	var count int
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM memberships WHERE workspace_id=$1 AND user_id=$2", w, target).Scan(&count))
	require.Equal(t, 1, count)
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM mutation_requests WHERE workspace_id=$1 AND request_key='removal-rollback-01'", w).Scan(&count))
	require.Zero(t, count)
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM audit_events WHERE workspace_id=$1", w).Scan(&count))
	require.Equal(t, auditsBefore, count, "failed commit never leaves protected audit")
}

func checkRemovalRace(t *testing.T, api string, db *fluxTesting.TestDB, token, other, w, actor, target string) {
	t.Helper()
	ctx := context.Background()
	lock, err := db.Pool.Begin(ctx)
	require.NoError(t, err)
	defer lock.Rollback(ctx)
	_, err = lock.Exec(ctx, "SELECT id FROM workspaces WHERE id=$1 FOR UPDATE", w)
	require.NoError(t, err)
	results := make(chan roleRequestResult, 2)
	go func() {
		results <- removalStatus(api, token, "/workspaces/"+w+"/members/"+actor, "removal-race-000001")
	}()
	go func() {
		results <- roleStatus(api, other, "/workspaces/"+w+"/members/"+target, "removal-race-000002", `{"role":"member"}`)
	}()
	require.Eventually(t, func() bool {
		var waiting int
		return db.Pool.QueryRow(ctx, "SELECT count(*) FROM pg_stat_activity WHERE datname=current_database() AND wait_event_type='Lock'").Scan(&waiting) == nil && waiting >= 2
	}, 2*time.Second, 10*time.Millisecond)
	require.NoError(t, lock.Commit(ctx))
	a, b := <-results, <-results
	require.NoError(t, a.err)
	require.NoError(t, b.err)
	require.ElementsMatch(t, []int{200, 409}, []int{a.code, b.code})
	var count int
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM memberships WHERE workspace_id=$1 AND role='owner'", w).Scan(&count))
	require.Equal(t, 1, count)
}

func removalStatus(api, token, path, key string) roleRequestResult {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, api+"/api/v1"+path, strings.NewReader(`{}`))
	if err != nil {
		return roleRequestResult{err: err}
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Origin", fixtureParty)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", key)
	response, err := (&http.Client{Timeout: 4 * time.Second}).Do(req)
	if err != nil {
		return roleRequestResult{err: err}
	}
	_, readErr := io.Copy(io.Discard, io.LimitReader(response.Body, 65536))
	return roleRequestResult{code: response.StatusCode, err: errors.Join(readErr, response.Body.Close())}
}

func checkTeamRoles(t *testing.T, api string, db *fluxTesting.TestDB, p *signedProvider) {
	t.Helper()
	ctx := context.Background()
	token := p.token(t, map[string]any{"sub": "user_roles", "sid": "sess_user_roles", "org_role": "org:owner"})
	other := p.token(t, map[string]any{"sub": "user_roles_other", "sid": "sess_user_roles_other"})
	status, created := workspaceRequest(t, api, token, "POST", "/workspaces", "roles-workspace-0001", `{"name":"Role tenant"}`)
	require.Equal(t, 201, status)
	w := created["workspace"].(map[string]any)["id"].(string)
	status, foreign := workspaceRequest(t, api, other, "POST", "/workspaces", "roles-workspace-0002", `{"name":"Foreign role tenant"}`)
	require.Equal(t, 201, status)
	f := foreign["workspace"].(map[string]any)["id"].(string)
	var actor, target string
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT user_id::text FROM memberships WHERE workspace_id=$1", w).Scan(&actor))
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT user_id::text FROM memberships WHERE workspace_id=$1", f).Scan(&target))
	path := "/workspaces/" + w + "/members/" + target
	status, _ = workspaceRequest(t, api, token, "PATCH", path, "roles-foreign-0001", `{"role":"member"}`)
	require.Equal(t, 404, status)
	_, err := db.Pool.Exec(ctx, "INSERT INTO memberships(workspace_id,user_id,role,created_by) VALUES($1,$2,'member',$3)", w, target, actor)
	require.NoError(t, err)
	checkTeamRoleMatrix(t, api, db, token, w, actor, target, path)
	_, err = db.Pool.Exec(ctx, "UPDATE memberships SET role=CASE WHEN user_id=$2 THEN 'owner' ELSE 'member' END WHERE workspace_id=$1", w, actor)
	require.NoError(t, err)
	self := "/workspaces/" + w + "/members/" + actor
	status, denied := workspaceRequest(t, api, token, "PATCH", self, "roles-last-owner-0001", `{"role":"admin"}`)
	require.Equal(t, 409, status)
	require.Equal(t, "OWNER_REQUIRED", denied["code"])
	for _, body := range []string{`{"role":"bogus"}`, `{"role":null}`, `{"role":"member","workspaceId":"` + f + `"}`, `null`, `{"role":"member"} {}`} {
		status, _ = workspaceRequest(t, api, token, "PATCH", path, "roles-invalid-0001", body)
		require.Equal(t, 400, status)
	}
	var auditsBefore int
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM audit_events WHERE workspace_id=$1", w).Scan(&auditsBefore))
	status, first := workspaceRequest(t, api, token, "PATCH", path, "roles-replay-0001", `{"role":"owner"}`)
	require.Equal(t, 200, status)
	require.Equal(t, "owner", first["actorRole"])
	status, replay := workspaceRequest(t, api, token, "PATCH", path, "roles-replay-0001", `{"role":"owner"}`)
	require.Equal(t, 200, status)
	require.Equal(t, first, replay)
	status, _ = workspaceRequest(t, api, token, "PATCH", path, "roles-replay-0001", `{"role":"admin"}`)
	require.Equal(t, 409, status)
	var count int
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM audit_events WHERE workspace_id=$1", w).Scan(&count))
	require.Equal(t, auditsBefore+1, count, "identical replay and conflicting reuse never audit twice")
	var validAudit bool
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM audit_events a JOIN memberships m ON "+
		"m.workspace_id=a.workspace_id AND m.id=a.target_membership_id WHERE a.workspace_id=$1 AND a.actor_id=$2 "+
		"AND a.operation='member.role' AND m.user_id=$3 AND a.created_at IS NOT NULL)", w, actor, target).Scan(&validAudit))
	require.True(t, validAudit)
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM mutation_requests WHERE workspace_id=$1 AND request_key='roles-replay-0001' AND operation='member.role' AND retain_until >= created_at+interval '24 hours'", w).Scan(&count))
	require.Equal(t, 1, count)
	status, _ = workspaceRequest(t, api, other, "PATCH", path, "roles-replay-0001", `{"role":"owner"}`)
	require.Equal(t, 200, status, "same key is scoped to its durable actor")
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM mutation_requests WHERE workspace_id=$1 AND operation='member.role' AND request_key='roles-replay-0001'", w).Scan(&count))
	require.Equal(t, 2, count)
	_, err = db.Pool.Exec(ctx, "INSERT INTO memberships(workspace_id,user_id,role,created_by) VALUES($1,$2,'owner',$2)", f, actor)
	require.NoError(t, err)
	status, _ = workspaceRequest(t, api, token, "PATCH", "/workspaces/"+f+"/members/"+target, "roles-replay-0001", `{"role":"owner"}`)
	require.Equal(t, 200, status, "same actor and key are scoped to their workspace")
	// Serialize both current owners behind a real workspace lock, then release.
	lock, err := db.Pool.Begin(ctx)
	require.NoError(t, err)
	defer lock.Rollback(ctx)
	_, err = lock.Exec(ctx, "SELECT id FROM workspaces WHERE id=$1 FOR UPDATE", w)
	require.NoError(t, err)
	results := make(chan roleRequestResult, 2)
	go func() {
		result := roleStatus(api, token, self, "roles-race-00001", `{"role":"member"}`)
		results <- result
	}()
	go func() {
		result := roleStatus(api, other, path, "roles-race-00002", `{"role":"member"}`)
		results <- result
	}()
	require.Eventually(t, func() bool {
		var waiting int
		return db.Pool.QueryRow(ctx, "SELECT count(*) FROM pg_stat_activity WHERE datname=current_database() AND wait_event_type='Lock'").Scan(&waiting) == nil && waiting >= 2
	}, 2*time.Second, 10*time.Millisecond)
	require.NoError(t, lock.Commit(ctx))
	a, b := <-results, <-results
	require.NoError(t, a.err)
	require.NoError(t, b.err)
	require.ElementsMatch(t, []int{200, 409}, []int{a.code, b.code})
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM memberships WHERE workspace_id=$1 AND role='owner'", w).Scan(&count))
	require.Equal(t, 1, count)
	_, err = db.Pool.Exec(ctx, "UPDATE memberships SET role=CASE WHEN user_id=$2 THEN 'owner' ELSE 'member' END WHERE workspace_id=$1", w, actor)
	require.NoError(t, err)
	checkRoleRollback(t, api, db, p, token, w, path, target)
	// A pending operation must see the actor role committed before its lock.
	lock, err = db.Pool.Begin(ctx)
	require.NoError(t, err)
	defer lock.Rollback(ctx)
	_, err = lock.Exec(ctx, "SELECT id FROM workspaces WHERE id=$1 FOR UPDATE", w)
	require.NoError(t, err)
	go func() {
		result := roleStatus(api, token, path, "roles-replay-0001", `{"role":"viewer"}`)
		results <- result
	}()
	require.Eventually(t, func() bool {
		var waiting int
		return db.Pool.QueryRow(ctx, "SELECT count(*) FROM pg_stat_activity WHERE datname=current_database() AND wait_event_type='Lock'").Scan(&waiting) == nil && waiting >= 1
	}, 2*time.Second, 10*time.Millisecond)
	_, err = lock.Exec(ctx, "UPDATE memberships SET role='member' WHERE workspace_id=$1 AND user_id=$2", w, actor)
	require.NoError(t, err)
	require.NoError(t, lock.Commit(ctx))
	pending := <-results
	require.NoError(t, pending.err)
	require.Equal(t, 403, pending.code, "fresh authorization precedes changed-hash replay")
	for _, grant := range []string{"owner", "admin", "member", "viewer"} {
		_, err = db.Pool.Exec(ctx, "UPDATE memberships SET role='admin' WHERE workspace_id=$1 AND user_id=$2", w, actor)
		require.NoError(t, err)
		status, _ = workspaceRequest(t, api, token, "PATCH", self, "roles-self-"+uuid.NewString(), `{"role":"`+grant+`"}`)
		require.Equal(t, 403, status, "admin cannot manage self or escalate")
	}
	_, err = db.Pool.Exec(ctx, "DELETE FROM memberships WHERE workspace_id=$1 AND user_id=$2", w, actor)
	require.NoError(t, err)
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM audit_events WHERE workspace_id=$1 AND actor_id=$2", w, actor).Scan(&count))
	require.Positive(t, count, "durable actor provenance survives membership removal")
	status, _ = workspaceRequest(t, api, token, "PATCH", path, "roles-replay-0001", `{"role":"owner"}`)
	require.Equal(t, 404, status)
	_, err = db.Pool.Exec(ctx, "UPDATE audit_events SET operation='tampered' WHERE workspace_id=$1", w)
	require.Error(t, err)
	_, err = db.Pool.Exec(ctx, "DELETE FROM audit_events WHERE workspace_id=$1", w)
	require.Error(t, err)
	require.NotContains(t, p.logs.String(), "Role tenant")
}

func checkTeamRoleMatrix(t *testing.T, api string, db *fluxTesting.TestDB, token, w, actor, target, path string) {
	t.Helper()
	ctx := context.Background()
	for _, actorRole := range []string{"owner", "admin", "member", "viewer"} {
		for _, current := range []string{"owner", "admin", "member", "viewer"} {
			for _, grant := range []string{"owner", "admin", "member", "viewer"} {
				t.Run(actorRole+"-"+current+"-"+grant, func(t *testing.T) {
					_, err := db.Pool.Exec(ctx, "UPDATE memberships SET role=CASE WHEN user_id=$2 THEN $3 ELSE $4 END WHERE workspace_id=$1", w, actor, actorRole, current)
					require.NoError(t, err)
					status, _ := workspaceRequest(t, api, token, "PATCH", path, "roles-matrix-"+uuid.NewString(), `{"role":"`+grant+`"}`)
					allowed := actorRole == "owner" || (actorRole == "admin" && (current == "member" || current == "viewer") && (grant == "member" || grant == "viewer"))
					expected := current
					if allowed {
						require.Equal(t, 200, status)
						expected = grant
					} else {
						require.Equal(t, 403, status)
					}
					var actual string
					require.NoError(t, db.Pool.QueryRow(ctx, "SELECT role FROM memberships WHERE workspace_id=$1 AND user_id=$2", w, target).Scan(&actual))
					require.Equal(t, expected, actual)
					time.Sleep(70 * time.Millisecond)
				})
			}
		}
	}
}

type roleRequestResult struct {
	err  error
	code int
}

// Goroutines return transport failures to the test goroutine; never require/Fatal here.
func roleStatus(api, token, path, key, body string) roleRequestResult {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPatch, api+"/api/v1"+path, strings.NewReader(body))
	if err != nil {
		return roleRequestResult{err: err}
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Origin", fixtureParty)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", key)
	response, err := (&http.Client{Timeout: 4 * time.Second}).Do(req)
	if err != nil {
		return roleRequestResult{err: err}
	}
	_, readErr := io.Copy(io.Discard, io.LimitReader(response.Body, 65536))
	return roleRequestResult{code: response.StatusCode, err: errors.Join(readErr, response.Body.Close())}
}

func checkRoleRollback(t *testing.T, api string, db *fluxTesting.TestDB, p *signedProvider, token, workspace, path, target string) {
	t.Helper()
	ctx := context.Background()
	var before int
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM audit_events WHERE workspace_id=$1", workspace).Scan(&before))
	_, err := db.Pool.Exec(ctx, `CREATE FUNCTION reject_role_commit() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'PRIVATE-ROLE-COMMIT'; END $$;
CREATE CONSTRAINT TRIGGER reject_role_commit AFTER UPDATE ON memberships DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION reject_role_commit()`)
	require.NoError(t, err)
	status, _ := workspaceRequest(t, api, token, "PATCH", path, "roles-rollback-0001", `{"role":"viewer"}`)
	require.Equal(t, 503, status)
	_, err = db.Pool.Exec(ctx, "DROP TRIGGER reject_role_commit ON memberships; DROP FUNCTION reject_role_commit()")
	require.NoError(t, err)
	var role string
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT role FROM memberships WHERE workspace_id=$1 AND user_id=$2", workspace, target).Scan(&role))
	require.Equal(t, "member", role)
	var count int
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM mutation_requests WHERE workspace_id=$1 AND request_key='roles-rollback-0001'", workspace).Scan(&count))
	require.Zero(t, count)
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM audit_events WHERE workspace_id=$1", workspace).Scan(&count))
	require.Equal(t, before, count)
	require.NotContains(t, p.logs.String(), "PRIVATE-ROLE-COMMIT")
}

func checkTeam(t *testing.T, api string, db *fluxTesting.TestDB, p *signedProvider) {
	t.Helper()
	ctx := context.Background()
	token := p.token(t, map[string]any{"sub": "user_team", "sid": "sess_user_team", "org_role": "org:owner"})
	foreignToken := p.token(t, map[string]any{"sub": "user_team_foreign", "sid": "sess_user_team_foreign"})
	code, created := workspaceRequest(t, api, token, "POST", "/workspaces", "team-workspace-0001", `{"name":"Team inspection"}`)
	require.Equal(t, 201, code)
	id := created["workspace"].(map[string]any)["id"].(string)
	code, foreign := workspaceRequest(t, api, foreignToken, "POST", "/workspaces", "team-workspace-0002", `{"name":"Foreign Team secret"}`)
	require.Equal(t, 201, code)
	foreignID := foreign["workspace"].(map[string]any)["id"].(string)
	path := "/workspaces/" + id + "/members"
	for _, role := range []string{"owner", "admin", "member", "viewer"} {
		t.Run(role, func(t *testing.T) {
			_, err := db.Pool.Exec(ctx, "UPDATE memberships SET role=$2 WHERE workspace_id=$1", id, role)
			require.NoError(t, err)
			status, response := workspaceRequest(t, api, token, "GET", path, "", "")
			if role == "owner" || role == "admin" {
				require.Equal(t, 200, status)
				items := response["items"].([]any)
				require.Len(t, items, 1)
				item := items[0].(map[string]any)
				require.Len(t, item, 4, "only safe identity, workspace and role")
				require.Equal(t, id, item["workspaceId"])
				require.Equal(t, role, item["role"])
				require.NotEmpty(t, item["email"])
			} else {
				require.Equal(t, 403, status, "provider org claims cannot grant Team")
			}
			status, denied := workspaceRequest(t, api, token, "GET", "/workspaces/"+foreignID+"/members", "", "")
			require.Equal(t, 404, status)
			bytes, err := json.Marshal(denied)
			require.NoError(t, err)
			require.NotContains(t, string(bytes), foreignID)
			require.NotContains(t, string(bytes), "Foreign Team secret")
		})
	}
	_, err := db.Pool.Exec(ctx, "UPDATE memberships SET role='owner' WHERE workspace_id=$1", id)
	require.NoError(t, err)
	for _, query := range []string{"?workspaceId=" + foreignID, "?role=owner", "?search=secret", "?after=bad", "?after=" + uuid.NewString() + "&after=" + uuid.NewString()} {
		code, _ = workspaceRequest(t, api, token, "GET", path+query, "", "")
		require.Equal(t, 400, code, query)
	}
	code, _ = workspaceRequest(t, api, token, "GET", path+"/"+uuid.NewString(), "", "")
	require.Equal(t, 405, code, "role endpoint never exposes a member through unsupported GET")
	// Fixed 25-row pages keep even worst-case JSON-escaped 320-character
	// identities below the browser's existing 64 KiB success boundary.
	_, err = db.Pool.Exec(ctx, "WITH seeded AS (INSERT INTO users(issuer,subject,verified_email) "+
		"SELECT $1,'team-page-'||n,repeat('a',306)||n||'@example.com' FROM generate_series(1,30) n RETURNING id) "+
		"INSERT INTO memberships(workspace_id,user_id,role,created_by) SELECT $2,id,'member',id FROM seeded", fixtureIssuer, id)
	require.NoError(t, err)
	code, listed := workspaceRequest(t, api, token, "GET", path, "", "")
	require.Equal(t, 200, code)
	require.Len(t, listed["items"].([]any), 25)
	encoded, err := json.Marshal(listed)
	require.NoError(t, err)
	require.Less(t, len(encoded), 65536)
	seen := map[string]bool{}
	for {
		for _, raw := range listed["items"].([]any) {
			item := raw.(map[string]any)
			require.Equal(t, id, item["workspaceId"])
			userID := item["id"].(string)
			require.False(t, seen[userID])
			seen[userID] = true
		}
		after, more := listed["nextAfter"].(string)
		if !more {
			break
		}
		code, listed = workspaceRequest(t, api, token, "GET", path+"?after="+after, "", "")
		require.Equal(t, 200, code)
	}
	require.Len(t, seen, 31)
	_, err = db.Pool.Exec(ctx, "DELETE FROM memberships WHERE workspace_id=$1", id)
	require.NoError(t, err)
	code, _ = workspaceRequest(t, api, token, "GET", path, "", "")
	require.Equal(t, 404, code, "removal before read must revoke access")
	logs := p.logs.String()
	require.NotContains(t, logs, "user_team")
	require.NotContains(t, logs, "Foreign Team secret")
	require.NotContains(t, logs, token)
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
	for _, query := range []string{"?limit=0", "?limit=101", "?limit=-1", "?limit=1.5", "?limit=abc", "?limit=1&limit=2", "?cursor=foreign", "?cursor=", "?workspaceId=" + uuid.NewString(), "?search=" + url.QueryEscape(strings.Repeat("x", 201)), "?state=unknown"} {
		code, _ := workspaceRequest(t, api, token, "GET", path+query, "", "")
		require.Equal(t, 400, code, query)
	}
	// A new head inserted between requests cannot shift an existing seek position.
	_, insertErr := db.Pool.Exec(ctx, "INSERT INTO links (workspace_id,creator_user_id,managed_host,short_key,destination,title,created_at) VALUES($1,$2,'go.flux.test','library-new-head','https://example.com/new','Inserted after first page','2026-02-01T12:00:00Z')", workspace, actor)
	require.NoError(t, insertErr)
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
	// A correctly signed token for a different effective filter must still fail.
	encoded, _, _ := strings.Cut(cursor, ".")
	payload, decodeErr := base64.RawURLEncoding.DecodeString(encoded)
	require.NoError(t, decodeErr)
	var cursorBody map[string]any
	require.NoError(t, json.Unmarshal(payload, &cursorBody))
	cursorBody["fingerprint"] = "different-effective-filter"
	payload, decodeErr = json.Marshal(cursorBody)
	require.NoError(t, decodeErr)
	encoded = base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, make([]byte, 32)) // Same explicit test-only fixture key.
	_, decodeErr = mac.Write([]byte(encoded))
	require.NoError(t, decodeErr)
	mismatch := encoded + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	mismatchStatus, mismatchError := workspaceRequest(t, api, token, "GET", path+"?cursor="+url.QueryEscape(mismatch), "", "")
	require.Equal(t, 400, mismatchStatus)
	require.Equal(t, "CURSOR_INVALID", mismatchError["code"])
	require.Equal(t, "This page is no longer available. Return to the first page.", mismatchError["message"])

	for _, invalid := range []string{"", "malformed", cursor[:len(cursor)-1] + "!", "A" + cursor[1:], strings.Repeat("x", 2049), strings.Repeat("x", 9000)} {
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
	status, denied := workspaceRequest(t, api, token, "GET", path+"?cursor="+url.QueryEscape(cursor), "", "")
	require.Equal(t, 404, status, "fresh membership precedes listing")
	require.NotContains(t, fmt.Sprint(denied), "Library 103")
}

func checkLinkSearch(t *testing.T, api string, db *fluxTesting.TestDB, p *signedProvider) {
	t.Helper()
	ctx := context.Background()
	token := p.token(t, nil)
	paths := make([]string, 2)
	workspaces := make([]string, 2)
	for i := range paths {
		code, body := workspaceRequest(t, api, token, "POST", "/workspaces", fmt.Sprintf("search-workspace-%02d", i), fmt.Sprintf(`{"name":"Search tenant %d"}`, i))
		require.Equal(t, 201, code)
		workspaces[i] = body["workspace"].(map[string]any)["id"].(string)
		paths[i] = "/workspaces/" + workspaces[i] + "/links"
	}
	var actor string
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT id::text FROM users WHERE subject='user_fixture'").Scan(&actor))
	titles := []string{"Literal 50% off", "Literal under_score", "Literal back\\slash", "Literal ' OR 1=1--", "Ordinary", "Title target"}
	states := []string{"active", "disabled", "archived", "deleted", "active", "active"}
	for tenant := range paths {
		for i, title := range titles {
			_, err := db.Pool.Exec(ctx, "INSERT INTO links(workspace_id,creator_user_id,managed_host,short_key,destination,title,lifecycle) VALUES($1,$2,'go.flux.test',$3,$4,$5,$6)", workspaces[tenant], actor, fmt.Sprintf("search-tenant-%d-key-%d", tenant, i), fmt.Sprintf("https://example.com/destination-target-%d", i), title, states[i])
			require.NoError(t, err)
		}
	}
	for _, tc := range []struct {
		search, state string
		count         int
	}{
		{"%", "nondeleted", 1}, {"_", "nondeleted", 1}, {`\`, "nondeleted", 1}, {"' OR 1=1--", "deleted", 1},
		{"Title TARGET", "active", 1}, {"destination-target-4", "active", 1}, {"search-tenant-0-key-4", "active", 1},
		{"", "nondeleted", 5}, {"", "active", 3}, {"", "disabled", 1}, {"", "archived", 1}, {"", "deleted", 1}, {"absent", "nondeleted", 0},
	} {
		code, body := workspaceRequest(t, api, token, "GET", paths[0]+"?search="+url.QueryEscape(tc.search)+"&state="+tc.state, "", "")
		require.Equal(t, 200, code, "search %q state %s", tc.search, tc.state)
		items := body["items"].([]any)
		require.Len(t, items, tc.count)
		for _, item := range items {
			require.Equal(t, workspaces[0], item.(map[string]any)["workspaceId"])
		}
	}
	code, first := workspaceRequest(t, api, token, "GET", paths[0]+"?limit=1&search=%20Literal%20&state=nondeleted", "", "")
	require.Equal(t, 200, code)
	cursor := first["nextCursor"].(string)
	code, second := workspaceRequest(t, api, token, "GET", paths[0]+"?limit=1&search=Literal&state=nondeleted&cursor="+url.QueryEscape(cursor), "", "")
	require.Equal(t, 200, code, "trimmed effective query must bind identically")
	require.NotEqual(t, first["items"].([]any)[0].(map[string]any)["id"], second["items"].([]any)[0].(map[string]any)["id"])
	for _, query := range []string{"?search=literal&state=nondeleted", "?search=Literal&state=active", "?search=Literalx&state=nondeleted"} {
		changedCode, changedBody := workspaceRequest(t, api, token, "GET", paths[0]+query+"&cursor="+url.QueryEscape(cursor), "", "")
		require.Equal(t, 400, changedCode)
		require.Equal(t, "CURSOR_INVALID", changedBody["code"])
	}
	code, body := workspaceRequest(t, api, token, "GET", paths[1]+"?search=Literal&cursor="+url.QueryEscape(cursor), "", "")
	require.Equal(t, 400, code)
	require.Equal(t, "CURSOR_INVALID", body["code"])
	code, foreignOnly := workspaceRequest(t, api, token, "GET", paths[1]+"?search=search-tenant-0-key", "", "")
	require.Equal(t, 200, code)
	require.Empty(t, foreignOnly["items"])
	for _, query := range []string{"?state=all", "?state=ACTIVE", "?state=", "?search=a&search=b", "?state=active&state=deleted", "?sort=title", "?search=" + url.QueryEscape(strings.Repeat("界", 201)), "?search=%FF", "?search=%00", "?search=" + url.QueryEscape(strings.Repeat(" ", 201))} {
		invalidCode, _ := workspaceRequest(t, api, token, "GET", paths[0]+query, "", "")
		require.Equal(t, 400, invalidCode, query)
	}
	code, _ = workspaceRequest(t, api, token, "GET", paths[0]+"?search="+url.QueryEscape(strings.Repeat("界", 200)), "", "")
	require.Equal(t, 200, code)
	_, err := db.Pool.Exec(ctx, "UPDATE memberships SET role='viewer' WHERE workspace_id=$1", workspaces[0])
	require.NoError(t, err)
	code, _ = workspaceRequest(t, api, token, "GET", paths[0]+"?search=Literal&state=deleted", "", "")
	require.Equal(t, 200, code)
	_, err = db.Pool.Exec(ctx, "DELETE FROM memberships WHERE workspace_id=$1", workspaces[0])
	require.NoError(t, err)
	code, body = workspaceRequest(t, api, token, "GET", paths[0]+"?search=Literal&cursor="+url.QueryEscape(cursor), "", "")
	require.Equal(t, 404, code)
	require.NotContains(t, fmt.Sprint(body), "Literal")
}

func checkInvitationExpiryRace(t *testing.T, api, token, path string) {
	t.Helper()
	var wait sync.WaitGroup
	expiryCodes := make(chan int, 2)
	for i := range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			expiryCodes <- concurrentInvitationRequest(api, token, path, fmt.Sprintf("invite-after-expiry-%02d", i))
		}()
	}
	wait.Wait()
	close(expiryCodes)
	var expiryResults []int
	for code := range expiryCodes {
		expiryResults = append(expiryResults, code)
	}
	require.ElementsMatch(t, []int{201, 409}, expiryResults)
}
