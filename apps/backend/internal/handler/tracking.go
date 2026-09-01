package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"flux/apps/backend/internal/model/analytics"
	"flux/apps/backend/internal/model/user"
	"flux/apps/backend/internal/repository"

	"github.com/labstack/echo/v4"
)

type TrackingWorkspaceResolver interface {
	GetWorkspaceByTrackingClientID(ctx context.Context, clientID string) (*user.Workspace, error)
}

type TrackingHandler struct {
	workspaceRepo TrackingWorkspaceResolver
	publisher     analytics.ConversionPublisher
}

func NewTrackingHandler(workspaceRepo TrackingWorkspaceResolver, publisher analytics.ConversionPublisher) *TrackingHandler {
	return &TrackingHandler{
		workspaceRepo: workspaceRepo,
		publisher:     publisher,
	}
}

func (h *TrackingHandler) TrackConversion(c echo.Context) error {
	clientID := c.QueryParam("client_id")
	if clientID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "client_id is required")
	}

	// 1. Resolve workspace securely
	workspace, err := h.workspaceRepo.GetWorkspaceByTrackingClientID(c.Request().Context(), clientID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) || err.Error() == "no rows in result set" {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid client_id")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to resolve workspace")
	}

	// 2. Parse payload
	var req analytics.TrackConversionRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid JSON payload")
	}

	// 3. Validate basics
	if req.ConversionID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "conversion_id is required")
	}
	if req.ConversionName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "conversion_name is required")
	}
	if len(req.ConversionName) > 255 {
		return echo.NewHTTPError(http.StatusBadRequest, "conversion_name is too long")
	}
	if len(req.ClickIDs) > 50 {
		return echo.NewHTTPError(http.StatusBadRequest, "too many click_ids provided")
	}

	// 4. Construct Event securely assigning WorkspaceID
	event := &analytics.ConversionEvent{
		ConversionID:   req.ConversionID,
		WorkspaceID:    workspace.ID.String(),
		Timestamp:      time.Now().UTC(),
		ConversionName: req.ConversionName,
		Revenue:        req.Revenue,
		Currency:       req.Currency,
		ClickIDs:       req.ClickIDs,
		VisitorID:      req.VisitorID,
	}
	if event.ClickIDs == nil {
		event.ClickIDs = []string{}
	}

	// 5. Publish
	err = h.publisher.PublishConversion(c.Request().Context(), event)
	if err != nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "failed to durably queue conversion")
	}

	return c.JSON(http.StatusAccepted, map[string]string{"status": "accepted"})
}
