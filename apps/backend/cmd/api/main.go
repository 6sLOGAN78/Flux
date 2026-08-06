// Package main provides the Echo v4 REST API server entrypoint for Flux backend.
package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"flux/apps/backend/internal/config"
	"flux/apps/backend/internal/database"
	customMiddleware "flux/apps/backend/internal/middleware"
	"flux/apps/backend/internal/modules/auth"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// HealthResponse represents the API health status payload.
type HealthResponse struct {
	Status   string `json:"status"`
	DBStatus string `json:"database"`
}

func main() {
	// Initialize Zerolog logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load application configuration")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	dbPool, dbErr := database.InitDBPool(ctx, cfg.DatabaseURL)
	cancel()

	if dbErr != nil {
		log.Warn().Err(dbErr).Msg("postgresql connection ping failed via pgx/v5")
	} else {
		log.Info().Msg("successfully connected and pinged postgresql database via pgx/v5")
		defer dbPool.Close()
	}

	authSvc := auth.NewAuthService(cfg.JWTSecret, cfg.ClerkSecretKey)

	e := echo.New()

	// Middlewares
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(customMiddleware.TracingMiddleware(cfg.NewRelicLicenseKey, "flux-backend"))

	// Static OpenAPI & Swagger UI routes
	e.Static("/static", "static")
	e.File("/docs", "static/openapi.html")

	// Public Health Check Endpoint using pgx/v5 ping
	e.GET("/health", func(c echo.Context) error {
		dbState := "connected"
		if dbPool == nil {
			dbState = "disconnected"
		} else {
			ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
			defer cancel()
			if err := dbPool.Ping(ctx); err != nil {
				dbState = "disconnected"
			}
		}

		return c.JSON(http.StatusOK, HealthResponse{
			Status:   "ok",
			DBStatus: dbState,
		})
	})

	// Protected API route group example using JWTMiddleware
	apiGroup := e.Group("/api/v1")
	apiGroup.Use(customMiddleware.JWTMiddleware(authSvc))
	apiGroup.GET("/me", func(c echo.Context) error {
		userID := c.Get("user_id")
		email := c.Get("email")
		return c.JSON(http.StatusOK, map[string]interface{}{
			"user_id": userID,
			"email":   email,
			"status":  "authenticated",
		})
	})

	log.Info().Msgf("starting flux Echo v4 api server on port %s...", cfg.ServerPort)
	if err := e.Start(":" + cfg.ServerPort); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("server terminated unexpectedly")
	}
}
