package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"flux/apps/backend/internal/errs"
	"flux/apps/backend/internal/model/campaign"
	"flux/apps/backend/internal/service"
)

type CampaignHandler struct {
	service *service.CampaignService
}

func NewCampaignHandler(service *service.CampaignService) *CampaignHandler {
	return &CampaignHandler{service: service}
}

func (h *CampaignHandler) CreateCampaign(c echo.Context) error {
	tenantIDRaw := c.Get("tenant_id")
	if tenantIDRaw == nil {
		return errs.NewAppError("UNAUTHORIZED", "missing workspace context", nil)
	}
	workspaceID, ok := tenantIDRaw.(uuid.UUID)
	if !ok {
		return errs.NewAppError("UNAUTHORIZED", "invalid workspace context type", nil)
	}

	var payload campaign.CreateCampaignPayload
	if err := c.Bind(&payload); err != nil {
		return errs.NewBadRequestError("invalid request payload", false, nil, nil, err)
	}

	camp, err := h.service.CreateCampaign(c.Request().Context(), workspaceID, &payload)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, camp)
}

func (h *CampaignHandler) GetCampaign(c echo.Context) error {
	tenantIDRaw := c.Get("tenant_id")
	if tenantIDRaw == nil {
		return errs.NewAppError("UNAUTHORIZED", "missing workspace context", nil)
	}
	workspaceID, ok := tenantIDRaw.(uuid.UUID)
	if !ok {
		return errs.NewAppError("UNAUTHORIZED", "invalid workspace context type", nil)
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.NewBadRequestError("invalid campaign id", false, nil, nil, err)
	}

	camp, err := h.service.GetCampaign(c.Request().Context(), workspaceID, id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, camp)
}

func (h *CampaignHandler) ListCampaigns(c echo.Context) error {
	tenantIDRaw := c.Get("tenant_id")
	if tenantIDRaw == nil {
		return errs.NewAppError("UNAUTHORIZED", "missing workspace context", nil)
	}
	workspaceID, ok := tenantIDRaw.(uuid.UUID)
	if !ok {
		return errs.NewAppError("UNAUTHORIZED", "invalid workspace context type", nil)
	}

	campaigns, err := h.service.ListCampaigns(c.Request().Context(), workspaceID)
	if err != nil {
		return err
	}
	
	if campaigns == nil {
		campaigns = []campaign.Campaign{}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"data": campaigns})
}

func (h *CampaignHandler) UpdateCampaign(c echo.Context) error {
	tenantIDRaw := c.Get("tenant_id")
	if tenantIDRaw == nil {
		return errs.NewAppError("UNAUTHORIZED", "missing workspace context", nil)
	}
	workspaceID, ok := tenantIDRaw.(uuid.UUID)
	if !ok {
		return errs.NewAppError("UNAUTHORIZED", "invalid workspace context type", nil)
	}

	var payload campaign.UpdateCampaignPayload
	if err := c.Bind(&payload); err != nil {
		return errs.NewBadRequestError("invalid request payload", false, nil, nil, err)
	}
	
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.NewBadRequestError("invalid campaign id", false, nil, nil, err)
	}
	payload.ID = id

	camp, err := h.service.UpdateCampaign(c.Request().Context(), workspaceID, &payload)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, camp)
}

func (h *CampaignHandler) DeleteCampaign(c echo.Context) error {
	tenantIDRaw := c.Get("tenant_id")
	if tenantIDRaw == nil {
		return errs.NewAppError("UNAUTHORIZED", "missing workspace context", nil)
	}
	workspaceID, ok := tenantIDRaw.(uuid.UUID)
	if !ok {
		return errs.NewAppError("UNAUTHORIZED", "invalid workspace context type", nil)
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.NewBadRequestError("invalid campaign id", false, nil, nil, err)
	}

	err = h.service.DeleteCampaign(c.Request().Context(), workspaceID, id)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
