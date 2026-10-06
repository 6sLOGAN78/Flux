package router

import (
	"github.com/6sLOGAN78/flux/internal/handler"

	"github.com/labstack/echo/v4"
)

func registerSystemRoutes(r *echo.Echo, h *handler.Handlers) {
	if h.Health != nil {
		RegisterHealthRoutes(r, h.Health)
	}

	r.GET("/static/openapi.json", h.OpenAPI.ServeOpenAPISpec)

	r.GET("/docs", h.OpenAPI.ServeOpenAPIUI)
}

// RegisterHealthRoutes installs the public system probes on any role listener.
func RegisterHealthRoutes(r *echo.Echo, health *handler.HealthHandler) {
	r.GET("/live", health.Live)
	r.GET("/ready", health.Ready)
}
