// Package v1 registers API v1 routes.
package v1

import (
	"net/http"

	"flux/apps/backend/internal/handler"
	customMiddleware "flux/apps/backend/internal/middleware"
	"flux/apps/backend/internal/service"

	"github.com/labstack/echo/v4"
)

// RegisterV1Routes attaches protected API v1 endpoints to Echo.
func RegisterV1Routes(e *echo.Echo, authSvc *service.AuthService, analyticsHandler *handler.AnalyticsHandler) {
	v1 := e.Group("/api/v1")
	v1.Use(customMiddleware.JWTMiddleware(authSvc))

	v1.GET("/me", func(c echo.Context) error {
		userID := c.Get("user_id")
		email := c.Get("email")
		return c.JSON(http.StatusOK, map[string]interface{}{
			"user_id": userID,
			"email":   email,
			"status":  "authenticated",
		})
	})

	if analyticsHandler != nil {
		v1.GET("/analytics/summary", analyticsHandler.GetSummary)
		v1.GET("/analytics/links/:id", analyticsHandler.GetLinkMetrics)
	}
}
