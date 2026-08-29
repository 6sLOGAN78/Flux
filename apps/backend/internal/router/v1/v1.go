package v1

import (
	"net/http"

	"flux/apps/backend/internal/handler"
	customMiddleware "flux/apps/backend/internal/middleware"
	"flux/apps/backend/internal/repository"

	"github.com/labstack/echo/v4"
)

// RegisterV1Routes attaches protected API v1 endpoints to Echo.
func RegisterV1Routes(e *echo.Echo, userRepo *repository.UserRepository, analyticsHandler *handler.AnalyticsHandler, linksHandler *handler.LinksHandler) {
	v1 := e.Group("/api/v1")
	
	// Protected routes
	protected := v1.Group("")
	protected.Use(customMiddleware.ClerkJWTMiddleware(userRepo))

	protected.GET("/me", func(c echo.Context) error {
		userID := c.Get("user_id")
		clerkID := c.Get("clerk_user_id")
		tenantID := c.Get("tenant_id")
		return c.JSON(http.StatusOK, map[string]interface{}{
			"user_id":       userID,
			"clerk_user_id": clerkID,
			"tenant_id":     tenantID,
			"status":        "authenticated",
		})
	})

	if analyticsHandler != nil {
		protected.GET("/analytics/summary", analyticsHandler.GetSummary)
		protected.GET("/analytics/links/:id", analyticsHandler.GetLinkMetrics)
	}

	if linksHandler != nil {
		links := protected.Group("/links")
		links.POST("", linksHandler.CreateLink)
		links.GET("", linksHandler.GetLinks)
		links.GET("/:id", linksHandler.GetLinkByID)
		links.PATCH("/:id", linksHandler.UpdateLink)
		links.DELETE("/:id", linksHandler.DeleteLink)
	}
}
