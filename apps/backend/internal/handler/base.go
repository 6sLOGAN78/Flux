// Package handler provides HTTP handlers for Echo routes.
package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// BaseHandler provides common utility methods for Echo handlers.
type BaseHandler struct{}

func (h *BaseHandler) Success(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}
