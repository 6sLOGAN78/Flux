// Package redirect provides the high-performance HTTP redirect engine and caching layer.
package redirect

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// RedirectHandler handles HTTP redirect requests.
type RedirectHandler struct {
	svc *RedirectService
}

// NewRedirectHandler initializes a RedirectHandler instance.
func NewRedirectHandler(svc *RedirectService) *RedirectHandler {
	return &RedirectHandler{svc: svc}
}

// HandleRedirect processes GET /:slug requests and performs HTTP 301/302 redirects.
func (h *RedirectHandler) HandleRedirect(c echo.Context) error {
	slug := c.Param("slug")
	if slug == "" {
		return echo.NewHTTPError(http.StatusNotFound, "slug parameter is required")
	}

	target, err := h.svc.ResolveRedirect(c.Request().Context(), slug)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "short link not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to resolve redirect")
	}

	// Validate link state
	if target.Status == "deleted" {
		return echo.NewHTTPError(http.StatusGone, "short link has been deleted")
	}

	if target.Status == "disabled" {
		return echo.NewHTTPError(http.StatusForbidden, "short link is disabled")
	}

	if target.ExpiresAt != nil && time.Now().After(*target.ExpiresAt) {
		return echo.NewHTTPError(http.StatusGone, "short link has expired")
	}

	if target.IsPasswordProtected {
		return echo.NewHTTPError(http.StatusUnauthorized, "password required to access short link")
	}

	code := target.RedirectCode
	if code != http.StatusMovedPermanently && code != http.StatusFound {
		code = http.StatusMovedPermanently
	}

	return c.Redirect(code, target.DestinationURL)
}
