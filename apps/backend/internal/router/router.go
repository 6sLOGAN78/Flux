package router

import (
	"net/http"
	"time"
	"flux/apps/backend/internal/handler"
	customMiddleware "flux/apps/backend/internal/middleware"
	"flux/apps/backend/internal/repository"
	v1 "flux/apps/backend/internal/router/v1"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

// InitRouter sets up system routes, API v1 routes, and the redirect engine route.
func InitRouter(e *echo.Echo, dbPool *pgxpool.Pool, userRepo *repository.UserRepository, redirectHandler *handler.RedirectHandler, analyticsHandler *handler.AnalyticsHandler, linksHandler *handler.LinksHandler, campaignHandler *handler.CampaignHandler, domainHandler *handler.DomainHandler, tlsAuthHandler *handler.TLSAuthHandler, trackingHandler *handler.TrackingHandler, limiterStore customMiddleware.LimiterStore) {
	RegisterSystemRoutes(e, dbPool)

	// Public routes (redirects and tracking)
	if redirectHandler != nil {
		e.GET("/:slug", redirectHandler.HandleRedirect)
	}

	if trackingHandler != nil && limiterStore != nil {
		// Public tracking endpoint.
		// Uses Redis sliding window rate limiter to protect Redis/ClickHouse from spam.
		limiter := customMiddleware.RateLimitMiddleware(limiterStore, 100, 1 * time.Minute)
		
		trackGroup := e.Group("/api/v1/events")
		// Explicitly allow CORS from any origin for this specific public tracking group
		trackGroup.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{http.MethodPost, http.MethodOptions},
			AllowHeaders: []string{echo.HeaderContentType},
		}))
		trackGroup.Use(limiter)
		trackGroup.POST("/track", trackingHandler.TrackConversion)
	}

	// Internal infrastructure routes
	if tlsAuthHandler != nil {
		internal := e.Group("/api/internal")
		internal.GET("/tls/ask", tlsAuthHandler.CheckAuthorization)
	}

	// V1 API Routes
	v1.RegisterV1Routes(e, userRepo, analyticsHandler, linksHandler, campaignHandler, domainHandler)
}
