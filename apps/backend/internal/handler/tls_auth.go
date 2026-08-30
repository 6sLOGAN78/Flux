package handler

import (
	"errors"
	"net/http"

	"flux/apps/backend/internal/lib/utils"
	"flux/apps/backend/internal/repository"
	"github.com/labstack/echo/v4"
)

type TLSAuthHandler struct {
	repo *repository.DomainRepository
	apiKey string
}

func NewTLSAuthHandler(repo *repository.DomainRepository, apiKey string) *TLSAuthHandler {
	return &TLSAuthHandler{
		repo: repo,
		apiKey: apiKey,
	}
}

// CheckAuthorization handles GET /api/internal/tls/ask?domain=...
func (h *TLSAuthHandler) CheckAuthorization(c echo.Context) error {
	// 1. Enforce Internal Authentication
	// Fail-closed: if the API key is not configured, we cannot authorize anyone.
	if h.apiKey == "" {
		return c.JSON(http.StatusUnauthorized, map[string]bool{"authorized": false})
	}

	token := c.Request().Header.Get("X-Internal-Token")
	if token != h.apiKey {
		return c.JSON(http.StatusUnauthorized, map[string]bool{"authorized": false})
	}

	// 2. Extract and normalize hostname
	domainQuery := c.QueryParam("domain")
	if domainQuery == "" {
		return c.JSON(http.StatusBadRequest, map[string]bool{"authorized": false})
	}

	hostname := utils.NormalizeHostname(domainQuery)
	if hostname == "" {
		return c.JSON(http.StatusBadRequest, map[string]bool{"authorized": false})
	}

	// 3. Database Check
	d, err := h.repo.GetDomainByHostname(c.Request().Context(), hostname)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.JSON(http.StatusOK, map[string]bool{"authorized": false}) // Safely deny unknown domains
		}
		// Log internal error but return safe denial
		c.Logger().Errorf("tls auth db error for domain %s: %v", hostname, err)
		return c.JSON(http.StatusInternalServerError, map[string]bool{"authorized": false})
	}

	// 4. Authorization Rules
	if d.Status == "active" {
		return c.JSON(http.StatusOK, map[string]bool{"authorized": true})
	}

	// Any other state (pending, verifying, failed, disabled) -> DENY
	return c.JSON(http.StatusOK, map[string]bool{"authorized": false})
}
