package testing_test

import (
	"flag"
	"net/http"
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
	"github.com/rs/zerolog"
)

//nolint:gochecknoglobals // Explicit test-binary mode never enters a production executable.
var browserFixture = flag.Bool("browser-fixture", false, "hold test-only loopback bearer fixture")

// TestBrowserProductFixture composes the production router in an external test
// package: reusable infrastructure must not import higher application layers.
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
		return router.NewRouter(srv, &handler.Handlers{OpenAPI: handler.NewOpenAPIHandler(srv)},
			&service.Services{Auth: auth, Identity: identity,
				Workspace: service.NewWorkspaceService(repository.NewWorkspaceRepository(db.Pool))})
	})
}
