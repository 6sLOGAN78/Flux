package handler

import (
	"fmt"
	"net/http"

	"github.com/6sLOGAN78/flux/internal/server"
	"github.com/6sLOGAN78/flux/static"

	"github.com/labstack/echo/v4"
)

// OpenAPIHandler serves embedded API documentation.
type OpenAPIHandler struct {
	Handler
}

// NewOpenAPIHandler constructs an embedded documentation handler.
func NewOpenAPIHandler(s *server.Server) *OpenAPIHandler {
	return &OpenAPIHandler{
		Handler: NewHandler(s),
	}
}

// ServeOpenAPIUI serves the embedded browser reference page.
func (h *OpenAPIHandler) ServeOpenAPIUI(c echo.Context) error {
	templateBytes, err := static.Assets.ReadFile("openapi.html")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Content-Security-Policy", static.DocumentationCSP)
	c.Response().Header().Set("X-Content-Type-Options", "nosniff")
	if err != nil {
		return fmt.Errorf("failed to read OpenAPI UI template: %w", err)
	}

	templateString := string(templateBytes)

	err = c.HTML(http.StatusOK, templateString)
	if err != nil {
		return fmt.Errorf("failed to write HTML response: %w", err)
	}

	return nil
}

// ServeOpenAPISpec serves the generated contract bytes without reserialization.
func (h *OpenAPIHandler) ServeOpenAPISpec(c echo.Context) error {
	data, err := static.Assets.ReadFile("openapi.json")
	if err != nil {
		return fmt.Errorf("failed to read OpenAPI document: %w", err)
	}
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("X-Content-Type-Options", "nosniff")
	if err50 := c.Blob(http.StatusOK, "application/json", data); err50 != nil {
		return fmt.Errorf("failed to write OpenAPI document: %w", err50)
	}
	return nil
}
