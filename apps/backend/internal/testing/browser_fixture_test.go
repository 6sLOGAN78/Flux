package testing_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/6sLOGAN78/flux/internal/app"
	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/database"
	"github.com/6sLOGAN78/flux/internal/handler"
	"github.com/6sLOGAN78/flux/internal/lib/invitationcrypto"
	"github.com/6sLOGAN78/flux/internal/repository"
	"github.com/6sLOGAN78/flux/internal/router"
	"github.com/6sLOGAN78/flux/internal/server"
	"github.com/6sLOGAN78/flux/internal/service"
	fluxTesting "github.com/6sLOGAN78/flux/internal/testing"
	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwks"
	"github.com/clerk/clerk-sdk-go/v2/session"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/rs/zerolog"
)

//nolint:gochecknoglobals // Explicit test-binary mode never enters a production executable.
var browserFixture = flag.Bool("browser-fixture", false, "hold test-only loopback bearer fixture")

// TestBrowserProductFixture composes the production router in an external test
// package: reusable infrastructure must not import higher application layers.
//
//nolint:lll,gocognit // Test-only actual API/worker composition retains exact SQL acknowledgement fences.
func TestBrowserProductFixture(t *testing.T) {
	fluxTesting.RunBrowserProductFixture(t, *browserFixture, func(cfg *config.Config,
		db *database.Database, p *fluxTesting.SignedProvider,
	) http.Handler {
		log := zerolog.Nop()
		cfg.Auth = config.AuthConfig{SecretKey: "test-only", Issuer: "https://fixture.clerk.accounts.dev",
			AuthorizedParties: []string{"http://127.0.0.1:3100"}}
		srv := &server.Server{Config: cfg, Logger: &log, DB: db}
		clients := &clerk.ClientConfig{BackendConfig: clerk.BackendConfig{
			URL: clerk.String(p.URL()), Key: clerk.String(cfg.Auth.SecretKey),
			HTTPClient: &http.Client{Timeout: 3 * time.Second}}}
		auth := service.NewAuthServiceWithClients(cfg.Auth, service.AuthClients{
			JWKS: jwks.NewClient(clients), Sessions: session.NewClient(clients), Users: user.NewClient(clients)})
		identity := service.NewIdentityService(repository.NewUserRepository(db.Pool), auth)
		cfg.Links = config.LinksConfig{CursorKey: base64.StdEncoding.EncodeToString(make([]byte, 32)),
			ManagedHost: "go.flux.test", BlockedHosts: []string{"blocked.example"}}
		workspace := service.NewWorkspaceService(repository.NewWorkspaceRepository(db.Pool))
		cfg.Invitations = config.InvitationConfig{ActiveKeyID: "fixture",
			EncryptionKeys: ` {"fixture":"` + base64.StdEncoding.EncodeToString(make([]byte, 32)) + `"}`,
			Sender:         "invites@example.test", PublicOrigin: "https://app.flux.test"}
		team := service.NewTeamService(repository.NewTeamRepository(db.Pool), workspace)
		queue, closeQueue := fluxTesting.SetupTestRedis(t)
		t.Cleanup(closeQueue)
		cfg.Redis = queue.Config
		cfg.Worker = config.RoleConfig{ListenAddress: "127.0.0.1:0", DrainTimeout: 3 * time.Second, ReadinessTimeout: time.Second}
		cfg.Observability = config.DefaultObservabilityConfig()
		var releaseOnce sync.Once
		release := make(chan struct{})
		var acknowledgements atomic.Int32
		var recoveryAttempts atomic.Int32
		provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var payload struct {
				HTML string   `json:"html"`
				To   []string `json:"to"`
			}
			if json.NewDecoder(r.Body).Decode(&payload) != nil || len(payload.To) != 1 {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			if payload.To[0] == "recovery@example.test" {
				recoveryAttempts.Add(1)
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			if payload.To[0] != "delivery@example.test" {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			select {
			case <-release:
			case <-r.Context().Done():
				return
			}
			acknowledgements.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"id": "fixture-ack"})
		}))
		t.Cleanup(provider.Close)
		worker1, err2 := app.NewWorkerWithInvitationSender(context.Background(), cfg, invitationFixtureSender{endpoint: provider.URL, client: provider.Client()})
		if err2 != nil {
			t.Fatal("browser invitation worker startup failed")
		}
		t.Cleanup(func() {
			if err3 := worker1.Close(context.Background()); err3 != nil {
				t.Error("browser invitation worker cleanup failed")
			}
		})
		product := router.NewRouter(srv, &handler.Handlers{OpenAPI: handler.NewOpenAPIHandler(srv)},
			&service.Services{Auth: auth, Identity: identity,
				Team:        team,
				Invitations: service.NewInvitationService(repository.NewInvitationRepository(db.Pool), team, cfg.Invitations),
				Workspace:   workspace, Links: service.NewLinkService(repository.NewLinkRepository(db.Pool), workspace, cfg.Links)})
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/__test/delivery-recovery" {
				if r.Method == http.MethodPost {
					// Simulate a long outage after an uncertain actual HTTP attempt.
					// Column discovery keeps RED on behavior, not missing migration SQL.
					var hasWindow bool
					readErr := db.Pool.QueryRow(r.Context(), "SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name='invitation_delivery_intents' AND column_name='dispatch_started_at')").Scan(&hasWindow)
					if readErr != nil {
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					query := `UPDATE invitation_delivery_intents d SET created_at=now()-interval '2 days',available_at=now() FROM invitations i WHERE d.workspace_id=i.workspace_id AND d.invitation_id=i.id AND d.workspace_id::text=$1 AND i.email='recovery@example.test' AND d.state='queued' AND d.attempts>=1`
					if hasWindow {
						query = `UPDATE invitation_delivery_intents d SET dispatch_started_at=now()-interval '2 days',available_at=now() FROM invitations i WHERE d.workspace_id=i.workspace_id AND d.invitation_id=i.id AND d.workspace_id::text=$1 AND i.email='recovery@example.test' AND d.state='queued' AND d.attempts>=1`
					}
					result, updateErr := db.Pool.Exec(r.Context(), query, r.URL.Query().Get("workspace"))
					if updateErr != nil || result.RowsAffected() != 1 {
						w.WriteHeader(http.StatusConflict)
						return
					}
					w.WriteHeader(http.StatusNoContent)
					return
				}
				var failed int
				readErr := db.Pool.QueryRow(r.Context(), `SELECT count(*) FROM invitation_delivery_intents d JOIN invitations i ON i.workspace_id=d.workspace_id AND i.id=d.invitation_id WHERE d.workspace_id::text=$1 AND i.email='recovery@example.test' AND d.state='failed' AND d.ciphertext IS NOT NULL AND coalesce((to_jsonb(d)->>'reconciliation_required')::boolean,false)`, r.URL.Query().Get("workspace")).Scan(&failed)
				if readErr != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				_ = json.NewEncoder(w).Encode(map[string]int{"providerAttempts": int(recoveryAttempts.Load()), "reconciliationBlocked": failed})
				return
			}
			if r.URL.Path == "/__test/delivery-release" && r.Method == http.MethodPost {
				releaseOnce.Do(func() { close(release) })
				w.WriteHeader(http.StatusNoContent)
				return
			}
			if r.URL.Path == "/__test/delivery-evidence" {
				var delivered int
				err4 := db.Pool.QueryRow(r.Context(), `SELECT count(*) FROM invitation_delivery_intents d JOIN invitations i ON i.workspace_id=d.workspace_id AND i.id=d.invitation_id WHERE d.workspace_id::text=$1 AND i.email='delivery@example.test' AND d.state='delivered' AND d.ciphertext IS NULL`, r.URL.Query().Get("workspace")).Scan(&delivered)
				if err4 != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				_ = json.NewEncoder(w).Encode(map[string]int{"acknowledgements": int(acknowledgements.Load()), "erasedDeliveries": delivered})
				return
			}
			product.ServeHTTP(w, r)
		})
	})
}

// Explicit test-only HTTP sender: no runtime environment can select it.
type invitationFixtureSender struct {
	client   *http.Client
	endpoint string
}

//nolint:lll // Keep scoped SQL fences, bounded validation and exact fixture assertions together.
func (s invitationFixtureSender) SendInvitation(ctx context.Context, message invitationcrypto.Message) error {
	payload, err5 := json.Marshal(map[string]any{"to": []string{message.To}, "from": message.From, "subject": message.Subject, "html": message.HTML})
	if err5 != nil {
		return err5
	}
	req, err5 := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint+"/emails", bytes.NewReader(payload))
	if err5 != nil {
		return err5
	}
	req.Header.Set("Idempotency-Key", message.IdempotencyKey)
	response, err5 := s.client.Do(req)
	if err5 != nil {
		return errors.New("fixture send failed")
	}
	defer response.Body.Close()
	var ack struct {
		ID string `json:"id"`
	}
	if response.StatusCode != http.StatusOK || json.NewDecoder(response.Body).Decode(&ack) != nil || ack.ID == "" {
		return errors.New("fixture acknowledgement failed")
	}
	return nil
}
