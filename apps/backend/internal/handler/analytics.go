package handler

import (
	"errors"
	"net/http"
	"strconv"

	"flux/apps/backend/internal/repository"

	"github.com/labstack/echo/v4"
)

type AnalyticsHandler struct {
	provider repository.AnalyticsProvider
}

func NewAnalyticsHandler(provider repository.AnalyticsProvider) *AnalyticsHandler {
	return &AnalyticsHandler{provider: provider}
}

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

	if h.provider == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "analytics provider not initialized")
	}

	summary, err := h.provider.GetSummary(c.Request().Context(), userID, page, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to retrieve analytics summary")
	}

	return c.JSON(http.StatusOK, summary)
}

func (h *AnalyticsHandler) GetLinkMetrics(c echo.Context) error {
	linkID := c.Param("id")
	if linkID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "link id parameter is required")
	}

	if h.provider == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "analytics provider not initialized")
	}

	metrics, err := h.provider.GetLinkMetrics(c.Request().Context(), linkID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "link metrics not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to retrieve link metrics")
	}

	return c.JSON(http.StatusOK, metrics)
}
