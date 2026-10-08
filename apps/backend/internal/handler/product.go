package handler

import (
	"net/http"
	"strings"

	"github.com/6sLOGAN78/flux/internal/errs"
	"github.com/6sLOGAN78/flux/internal/service"
	"github.com/6sLOGAN78/flux/internal/transport"
	"github.com/labstack/echo/v4"
)

// ProductHandler adapts verified control-plane identity to canonical transports.
type ProductHandler struct{ auth *service.AuthService }

// NewProductHandler injects the authentication boundary explicitly.
func NewProductHandler(auth *service.AuthService) *ProductHandler { return &ProductHandler{auth: auth} }

// Me proves authentication without fabricating a durable Flux user identifier.
func (h *ProductHandler) Me(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store")
	values := c.Request().Header.Values("Authorization")
	if len(values) != 1 {
		return errs.NewUnauthorizedError("Authentication required", false)
	}
	scheme, token, ok := strings.Cut(values[0], " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return errs.NewUnauthorizedError("Authentication required", false)
	}
	_, err := h.auth.Authenticate(c.Request().Context(), token)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, transport.TransportIdentityResponse{Authenticated: true})
}
