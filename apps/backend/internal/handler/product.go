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
type ProductHandler struct {
	auth     *service.AuthService
	identity service.IdentityResolver
}

// NewProductHandler injects the authentication boundary explicitly.
func NewProductHandler(auth *service.AuthService, identity service.IdentityResolver) *ProductHandler {
	return &ProductHandler{auth: auth, identity: identity}
}

// Me returns the committed internal identity for an actively verified bearer.
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
	actor, err := h.auth.Authenticate(c.Request().Context(), token)
	if err != nil {
		return err
	}
	if h.identity == nil {
		return &errs.HTTPError{Code: "SERVICE_UNAVAILABLE", Message: "Authentication temporarily unavailable",
			Status: http.StatusServiceUnavailable}
	}
	user, err := h.identity.Resolve(c.Request().Context(), actor)
	if err != nil {
		return err
	}
	response := transport.TransportIdentityResponse{Authenticated: true}
	response.User.Id = user.ID.String()
	response.User.Email = user.Email
	return c.JSON(http.StatusOK, response)
}
