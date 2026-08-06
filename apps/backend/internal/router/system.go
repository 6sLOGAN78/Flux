// Package router registers system and API v1 endpoints.
package router

import (
	"flux/apps/backend/internal/handler"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// RegisterSystemRoutes registers health check and OpenAPI static documentation routes.
func RegisterSystemRoutes(e *echo.Echo, dbPool *pgxpool.Pool) {
	e.Static("/static", "static")
	e.File("/docs", "static/openapi.html")

	healthHandler := handler.NewHealthHandler(dbPool)
	e.GET("/health", healthHandler.HealthCheck)
}
