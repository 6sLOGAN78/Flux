package handler

import (
	"net/http"
	"time"

	"flux/apps/backend/internal/modules/attribution"
	"flux/apps/backend/internal/repository"

	"github.com/labstack/echo/v4"
)

func (h *AnalyticsHandler) GetAttribution(c echo.Context) error {
	workspaceID, err := h.getTenantID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid tenant")
	}

	from, to, err := parseDateRange(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	h.enforceRetention(c, workspaceID, &from)

	modelParam := c.QueryParam("model")
	if modelParam == "" {
		modelParam = string(attribution.ModelLinear)
	}

	attrModel := attribution.AttributionModel(modelParam)
	switch attrModel {
	case attribution.ModelFirstTouch, attribution.ModelLastTouch, attribution.ModelLinear, attribution.ModelTimeDecay, attribution.ModelPositionBased:
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "unsupported attribution model")
	}

	attrProvider, ok := h.provider.(repository.AttributionProvider)
	if !ok {
		return echo.NewHTTPError(http.StatusNotImplemented, "attribution not supported by current analytical provider")
	}

	conversions, err := attrProvider.GetConversionsWithTouchpoints(c.Request().Context(), workspaceID, from, to)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to query conversions")
	}

	// Empty state
	if len(conversions) == 0 {
		return c.JSON(http.StatusOK, &attribution.AttributionResult{
			Model:                  attrModel,
			TotalConversions:       0,
			TotalAttributedRevenue: 0,
			Campaigns:              []attribution.CampaignAttribution{},
		})
	}

	calculator := attribution.NewCalculator()
	result, err := calculator.Calculate(conversions, attrModel, 7*24*time.Hour)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "attribution calculation failed")
	}

	if result.Campaigns == nil {
		result.Campaigns = []attribution.CampaignAttribution{}
	}

	return c.JSON(http.StatusOK, result)
}
