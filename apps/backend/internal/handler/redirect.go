package handler

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"flux/apps/backend/internal/model/analytics"
	"flux/apps/backend/internal/repository"
	"flux/apps/backend/internal/service"

	"github.com/google/uuid"
	"flux/apps/backend/internal/lib/utils"
	"github.com/labstack/echo/v4"
)

type RedirectHandler struct {
	svc       *service.RedirectService
	publisher analytics.AnalyticsPublisher
}

func NewRedirectHandler(svc *service.RedirectService, publisher analytics.AnalyticsPublisher) *RedirectHandler {
	return &RedirectHandler{
		svc:       svc,
		publisher: publisher,
	}
}

func (h *RedirectHandler) HandleRedirect(c echo.Context) error {
	slug := c.Param("slug")
	if slug == "" {
		return echo.NewHTTPError(http.StatusNotFound, "slug parameter is required")
	}

	hostname := utils.NormalizeHostname(c.Request().Host)

	target, err := h.svc.ResolveRedirect(c.Request().Context(), hostname, slug)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "short link not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to resolve redirect")
	}

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

	destination := target.DestinationURL
	eventID := uuid.New().String()

	if target.Status == "active" {
		// 13B: Decorate destination URL with flux_cid
		if parsedURL, err := url.Parse(destination); err == nil {
			q := parsedURL.Query()
			q.Set("flux_cid", eventID)
			parsedURL.RawQuery = q.Encode()
			destination = parsedURL.String()
		}

		// Generate Analytics Event Non-Blockingly
		if h.publisher != nil {
			event := &analytics.AnalyticsEvent{
				EventID:        eventID,
				EventType:      analytics.EventTypeLinkRedirect,
				Timestamp:      time.Now().UTC(),
				LinkID:         target.LinkID,
				WorkspaceID:    target.TenantID,
				ShortCode:      target.Slug,
				CampaignID:     target.CampaignID,
				UTMSource:      target.UTMSource,
				UTMMedium:      target.UTMMedium,
				UTMCampaign:    target.UTMCampaign,
				UTMTerm:        target.UTMTerm,
				UTMContent:     target.UTMContent,
				CustomDomainID: target.CustomDomainID,
				Hostname:       target.Hostname,
				Referrer:       c.Request().Referer(),
				UserAgent:      c.Request().UserAgent(),
				IPHash:         utils.HashIP(c.RealIP()),
			}

			// Publish non-blockingly to the bounded queue
			_ = h.publisher.PublishEvent(c.Request().Context(), event)
		}
	}

	return c.Redirect(code, destination)
}
