package analytics

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

var (
	ErrLinkMetricsNotFound = errors.New("link metrics not found")
)

// AnalyticsSummaryResponse represents aggregated user dashboard statistics.
type AnalyticsSummaryResponse struct {
	TotalClicks  int64            `json:"total_clicks"`
	UniqueUsers  int64            `json:"unique_users"`
	TopCountries map[string]int64 `json:"top_countries"`
	TopBrowsers  map[string]int64 `json:"top_browsers"`
	TopDevices   map[string]int64 `json:"top_devices"`
	Page         int              `json:"page"`
	Limit        int              `json:"limit"`
}

// LinkMetricsResponse represents analytics breakdown for a single short link.
type LinkMetricsResponse struct {
	LinkID      string           `json:"link_id"`
	TotalClicks int64            `json:"total_clicks"`
	DailyStats  map[string]int64 `json:"daily_stats"`
}

// AnalyticsProvider defines querying capabilities for analytics data.
type AnalyticsProvider interface {
	GetSummary(ctx context.Context, userID string, page, limit int) (*AnalyticsSummaryResponse, error)
	GetLinkMetrics(ctx context.Context, linkID string) (*LinkMetricsResponse, error)
}

// AnalyticsHandler handles REST API requests for analytics metrics.
type AnalyticsHandler struct {
	provider AnalyticsProvider
}

// NewAnalyticsHandler initializes an AnalyticsHandler instance.
func NewAnalyticsHandler(provider AnalyticsProvider) *AnalyticsHandler {
	return &AnalyticsHandler{provider: provider}
}

// RegisterRoutes attaches analytics endpoints to an Echo group.
func (h *AnalyticsHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/analytics/summary", h.GetSummary)
	g.GET("/analytics/links/:id", h.GetLinkMetrics)
}

// GetSummary returns aggregated dashboard analytics with pagination parameters.
func (h *AnalyticsHandler) GetSummary(c echo.Context) error {
	userID, _ := c.Get("user_id").(string)

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page <= 0 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	summary, err := h.provider.GetSummary(c.Request().Context(), userID, page, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to retrieve analytics summary")
	}

	return c.JSON(http.StatusOK, summary)
}

// GetLinkMetrics returns analytics metrics for a specific link ID.
func (h *AnalyticsHandler) GetLinkMetrics(c echo.Context) error {
	linkID := c.Param("id")
	if linkID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "link id parameter is required")
	}

	metrics, err := h.provider.GetLinkMetrics(c.Request().Context(), linkID)
	if err != nil {
		if errors.Is(err, ErrLinkMetricsNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "link metrics not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to retrieve link metrics")
	}

	return c.JSON(http.StatusOK, metrics)
}
