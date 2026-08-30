package router

import (
	"flux/apps/backend/internal/handler"
	"flux/apps/backend/internal/repository"
	v1 "flux/apps/backend/internal/router/v1"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// InitRouter sets up system routes, API v1 routes, and the redirect engine route.
func InitRouter(e *echo.Echo, dbPool *pgxpool.Pool, userRepo *repository.UserRepository, redirectHandler *handler.RedirectHandler, analyticsHandler *handler.AnalyticsHandler, linksHandler *handler.LinksHandler, campaignHandler *handler.CampaignHandler, domainHandler *handler.DomainHandler, tlsAuthHandler *handler.TLSAuthHandler) {
	RegisterSystemRoutes(e, dbPool)

	// Public routes (redirects and tracking)
	if redirectHandler != nil {
		e.GET("/:slug", redirectHandler.HandleRedirect)
	}

	// Internal infrastructure routes
	if tlsAuthHandler != nil {
		internal := e.Group("/api/internal")
		internal.GET("/tls/ask", tlsAuthHandler.CheckAuthorization)
	}

	// V1 API Routes
	v1.RegisterV1Routes(e, userRepo, analyticsHandler, linksHandler, campaignHandler, domainHandler)
}
