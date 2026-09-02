package router

import (
	"flux/apps/backend/internal/handler"

	"github.com/labstack/echo/v4"
)

// RegisterWebhookRoutes attaches public webhook endpoints to Echo.
func RegisterWebhookRoutes(e *echo.Echo, stripeWebhookHandler *handler.StripeWebhookHandler) {
	if stripeWebhookHandler != nil {
		e.POST("/api/v1/webhooks/stripe", stripeWebhookHandler.HandleWebhook)
	}
}
