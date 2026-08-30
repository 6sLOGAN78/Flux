package handler

import (
	"net/http"
	"strconv"
	"time"

	"flux/apps/backend/internal/repository"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type AnalyticsHandler struct {
	provider repository.AnalyticsProvider
}

func NewAnalyticsHandler(provider repository.AnalyticsProvider) *AnalyticsHandler {
	return &AnalyticsHandler{provider: provider}
}

func (h *AnalyticsHandler) getTenantID(c echo.Context) (string, error) {
	tenantID, ok := c.Get("tenant_id").(uuid.UUID)
	if !ok {
		return "", echo.NewHTTPError(http.StatusUnauthorized, "workspace context not found")
	}
	return tenantID.String(), nil
}

// parseDateRange parses `from` and `to` query parameters, applying sensible defaults and limits.
func parseDateRange(c echo.Context) (time.Time, time.Time, error) {
	now := time.Now().UTC()
	to := now
	from := now.AddDate(0, -1, 0) // Default: last 30 days

	if toStr := c.QueryParam("to"); toStr != "" {
		parsed, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			return from, to, echo.NewHTTPError(http.StatusBadRequest, "invalid 'to' date format, expected RFC3339")
		}
		to = parsed.UTC()
	}

	if fromStr := c.QueryParam("from"); fromStr != "" {
		parsed, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			return from, to, echo.NewHTTPError(http.StatusBadRequest, "invalid 'from' date format, expected RFC3339")
		}
		from = parsed.UTC()
	}

	if from.After(to) {
		return from, to, echo.NewHTTPError(http.StatusBadRequest, "'from' date must be before 'to' date")
	}

	// Prevent unbounded expensive queries (max 1 year)
	if to.Sub(from) > 366*24*time.Hour {
		return from, to, echo.NewHTTPError(http.StatusBadRequest, "date range cannot exceed 1 year")
	}

	return from, to, nil
}

func (h *AnalyticsHandler) GetSummary(c echo.Context) error {
	tenantID, err := h.getTenantID(c)
	if err != nil {
		return err
	}

	from, to, err := parseDateRange(c)
	if err != nil {
		return err
	}

	if h.provider == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "analytics provider not initialized")
	}

	summary, err := h.provider.GetSummary(c.Request().Context(), tenantID, from, to)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to retrieve analytics summary")
	}

	return c.JSON(http.StatusOK, summary)
}

func (h *AnalyticsHandler) GetTimeseries(c echo.Context) error {
	tenantID, err := h.getTenantID(c)
	if err != nil {
		return err
	}

	from, to, err := parseDateRange(c)
	if err != nil {
		return err
	}

	interval := c.QueryParam("interval")
	if interval != "hour" && interval != "day" {
		interval = "day" // default
	}

	if h.provider == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "analytics provider not initialized")
	}

	timeseries, err := h.provider.GetTimeseries(c.Request().Context(), tenantID, from, to, interval)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to retrieve analytics timeseries")
	}

	return c.JSON(http.StatusOK, timeseries)
}

func (h *AnalyticsHandler) GetTopLinks(c echo.Context) error {
	tenantID, err := h.getTenantID(c)
	if err != nil {
		return err
	}

	from, to, err := parseDateRange(c)
	if err != nil {
		return err
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	if h.provider == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "analytics provider not initialized")
	}

	topLinks, err := h.provider.GetTopLinks(c.Request().Context(), tenantID, from, to, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to retrieve top links")
	}

	return c.JSON(http.StatusOK, topLinks)
}

func (h *AnalyticsHandler) GetReferrers(c echo.Context) error {
	tenantID, err := h.getTenantID(c)
	if err != nil {
		return err
	}

	from, to, err := parseDateRange(c)
	if err != nil {
		return err
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	if h.provider == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "analytics provider not initialized")
	}

	referrers, err := h.provider.GetReferrers(c.Request().Context(), tenantID, from, to, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to retrieve referrers")
	}

	return c.JSON(http.StatusOK, referrers)
}

// Deprecated link-specific endpoint that must be scoped by tenant ID and link ID
func (h *AnalyticsHandler) GetLinkMetrics(c echo.Context) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "use standard workspace-scoped /analytics endpoints")
}

func (h *AnalyticsHandler) GetCampaignPerformance(c echo.Context) error {
	tenantID, err := h.getTenantID(c)
	if err != nil {
		return err
	}

	from, to, err := parseDateRange(c)
	if err != nil {
		return err
	}

	limit := 50
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	perf, err := h.provider.GetCampaignPerformance(c.Request().Context(), tenantID, from, to, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to retrieve campaign performance")
	}

	return c.JSON(http.StatusOK, perf)
}

func (h *AnalyticsHandler) GetUTMPerformance(c echo.Context) error {
	tenantID, err := h.getTenantID(c)
	if err != nil {
		return err
	}

	from, to, err := parseDateRange(c)
	if err != nil {
		return err
	}

	limit := 50
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	dimension := c.QueryParam("dimension")
	if dimension == "" {
		dimension = "utm_source"
	}

	validDimensions := map[string]bool{
		"utm_source":   true,
		"utm_medium":   true,
		"utm_campaign": true,
		"utm_term":     true,
		"utm_content":  true,
	}

	if !validDimensions[dimension] {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid UTM dimension")
	}

	perf, err := h.provider.GetUTMPerformance(c.Request().Context(), tenantID, dimension, from, to, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to retrieve UTM performance")
	}

	return c.JSON(http.StatusOK, perf)
}

// GetDomainPerformance handles GET /api/v1/analytics/domains
func (h *AnalyticsHandler) GetDomainPerformance(c echo.Context) error {
	tenantID, err := h.getTenantID(c)
	if err != nil {
		return err
	}

	from, to, err := parseDateRange(c)
	if err != nil {
		return err
	}

	limitStr := c.QueryParam("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	resp, err := h.provider.GetDomainPerformance(c.Request().Context(), tenantID, from, to, limit)
	if err != nil {
		c.Logger().Errorf("failed to get domain performance: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get domain performance")
	}

	return c.JSON(http.StatusOK, resp)
}
