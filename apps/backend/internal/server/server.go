package server

import (
	"context"
	"net/http"
	"time"

	"flux/apps/backend/internal/config"
	"flux/apps/backend/internal/database"
	"flux/apps/backend/internal/handler"
	"flux/apps/backend/internal/logger"
	customMiddleware "flux/apps/backend/internal/middleware"
	"flux/apps/backend/internal/repository"
	"flux/apps/backend/internal/router"
	"flux/apps/backend/internal/service"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// Server represents the API server instance.
type Server struct {
	Echo   *echo.Echo
	Config *config.Config
	DBPool *pgxpool.Pool
}

// NewServer initializes and wires all dependencies for the server.
func NewServer(cfg *config.Config) (*Server, error) {
	logger.InitLogger("debug", "console")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	dbPool, dbErr := database.InitDBPool(ctx, cfg.DatabaseURL)
	cancel()

	if dbErr != nil {
		log.Warn().Err(dbErr).Msg("postgresql connection ping failed via pgx/v5")
	} else {
		log.Info().Msg("successfully connected and pinged postgresql database via pgx/v5")
		migCtx, migCancel := context.WithTimeout(context.Background(), 10*time.Second)
		if migErr := database.MigrateDSN(migCtx, &log.Logger, cfg.DatabaseURL); migErr != nil {
			log.Warn().Err(migErr).Msg("database migration warning")
		}
		migCancel()
	}

	if cfg.ClerkSecretKey != "" {
		clerk.SetKey(cfg.ClerkSecretKey)
	} else {
		log.Warn().Msg("CLERK_SECRET_KEY is empty, authentication may fail")
	}

	redirectRepo := repository.NewPostgresRedirectRepository(dbPool)
	redirectSvc := service.NewRedirectService(redirectRepo, nil)
	redirectHandler := handler.NewRedirectHandler(redirectSvc)
	analyticsHandler := handler.NewAnalyticsHandler(nil)
	
	linkRepo := repository.NewLinkRepository(dbPool)
	linkSvc := service.NewLinkService(linkRepo)
	linksHandler := handler.NewLinksHandler(linkSvc)

	userRepo := repository.NewUserRepository(dbPool)

	e := echo.New()
	customMiddleware.RegisterGlobalMiddlewares(e)
	e.Use(customMiddleware.TracingMiddleware(cfg.NewRelicLicenseKey, "flux-backend"))

	router.InitRouter(e, dbPool, userRepo, redirectHandler, analyticsHandler, linksHandler)

	return &Server{
		Echo:   e,
		Config: cfg,
		DBPool: dbPool,
	}, nil
}

// Start launches the Echo HTTP listener.
func (s *Server) Start() error {
	log.Info().Msgf("starting flux Echo v4 api server on port %s...", s.Config.ServerPort)
	if err := s.Echo.Start(":" + s.Config.ServerPort); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
