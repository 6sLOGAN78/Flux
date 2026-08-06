package router

import (
	"flux/apps/backend/internal/handler"
	v1 "flux/apps/backend/internal/router/v1"
	"flux/apps/backend/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// InitRouter sets up system routes, API v1 routes, and the redirect engine route.
func InitRouter(e *echo.Echo, dbPool *pgxpool.Pool, authSvc *service.AuthService, redirectHandler *handler.RedirectHandler, analyticsHandler *handler.AnalyticsHandler) {
	RegisterSystemRoutes(e, dbPool)
	v1.RegisterV1Routes(e, authSvc, analyticsHandler)

	if redirectHandler != nil {
		e.GET("/:slug", redirectHandler.HandleRedirect)
	}
}
