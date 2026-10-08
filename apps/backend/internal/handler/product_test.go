package handler_test

import (
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
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

const fixtureIssuer = "https://fixture.clerk.accounts.dev"
const fixtureParty = "http://127.0.0.1:3100"

type signedProvider struct {
	key         *rsa.PrivateKey
	server      *httptest.Server
	status      string
	mu          sync.Mutex
	failure     bool
	requests    int
	profileMode string
	profiles    int
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
		if strings.HasPrefix(r.URL.Path, "/users/") {
			id := strings.TrimPrefix(r.URL.Path, "/users/")
			p.profiles++
			verified := "verified"
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
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"id": id, "primary_email_address_id": "email_fixture",
				"email_addresses": []any{map[string]any{"id": "email_fixture", "email_address": "same@example.test",
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

func productRouter(cfg *config.Config, db *database.Database, p *signedProvider) http.Handler {
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

// TestProductActualHTTP is the registered real-PostgreSQL product dispatcher.
func TestProductActualHTTP(t *testing.T) {
	db, cleanup := fluxTesting.SetupTestDB(t)
	defer cleanup()
	p := newSignedProvider(t)
	api := httptest.NewServer(productRouter(db.Config, &database.Database{Pool: db.Pool}, p))
	defer api.Close()
	token := p.token(t, nil)
	t.Run("identity-stable-committed-UUID", func(t *testing.T) {
		status, body, cache := requestMe(t, api.URL, token, false)
		require.Equal(t, 200, status)
		require.Equal(t, "no-store", cache)
		var identity struct{ User struct{ ID string } }
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
	})
	for _, tc := range []struct {
		mode   string
		status int
	}{
		{"failure", 503}, {"slow", 503}, {"mismatch", 401}, {"unverified", 401},
	} {
		t.Run("identity-profile-"+tc.mode, func(t *testing.T) {
			provider := newSignedProvider(t)
			provider.profileMode = tc.mode
			api := httptest.NewServer(productRouter(db.Config, &database.Database{Pool: db.Pool}, provider))
			defer api.Close()
			subject := "user_" + tc.mode
			started := time.Now()
			status, body, cache := requestMe(t, api.URL, provider.token(t, map[string]any{"sub": subject, "sid": "sess_" + subject}), false)
			require.Equal(t, tc.status, status)
			require.Equal(t, "no-store", cache)
			require.Less(t, time.Since(started), 3*time.Second)
			for _, secret := range []string{subject, fixtureIssuer, "same@example.test", "PRIVATE-PROFILE-MARKER"} {
				require.NotContains(t, body, secret)
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
			_, err := db.Pool.Exec(context.Background(), "DROP TRIGGER reject_identity_commit ON users; DROP FUNCTION reject_identity_commit()")
			require.NoError(t, err)
		}()
		status, body, cache := requestMe(t, api.URL, p.token(t, map[string]any{"sub": "user_rollback", "sid": "sess_user_rollback"}), false)
		require.Equal(t, 503, status)
		require.Equal(t, "no-store", cache)
		require.NotContains(t, body, "PRIVATE-COMMIT-MARKER")
		var rows int
		require.NoError(t, db.Pool.QueryRow(context.Background(), "SELECT count(*) FROM users WHERE subject=$1", "user_rollback").Scan(&rows))
		require.Zero(t, rows)
	})
}
