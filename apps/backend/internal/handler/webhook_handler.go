package handler

import (
	"net/http"

	"flux/apps/backend/internal/model/webhook"
	"flux/apps/backend/internal/repository"
	"flux/apps/backend/internal/lib/utils"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type WebhookHandler struct {
	repo *repository.WebhookRepository
}

func NewWebhookHandler(repo *repository.WebhookRepository) *WebhookHandler {
	return &WebhookHandler{repo: repo}
}

func (h *WebhookHandler) CreateWebhook(c echo.Context) error {
	workspaceIDStr := c.Get("tenant_id").(string)
	if workspaceIDStr == "" {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "tenant_id missing"})
	}
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid tenant_id"})
	}

	var req webhook.CreateWebhookPayload
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if req.EndpointURL == "" || len(req.Events) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "endpoint_url and events are required"})
	}

	secret, err := utils.GenerateWebhookSecret()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate webhook secret"})
	}

	wh := &webhook.Webhook{
		WorkspaceID: workspaceID,
		EndpointURL: req.EndpointURL,
		Events:      req.Events,
		Active:      true,
		Secret:      secret,
	}

	created, err := h.repo.CreateWebhook(c.Request().Context(), wh)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Make sure secret is included exactly once on creation
	return c.JSON(http.StatusCreated, created)
}

func (h *WebhookHandler) ListWebhooks(c echo.Context) error {
	workspaceIDStr := c.Get("tenant_id").(string)
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid tenant_id"})
	}

	webhooks, err := h.repo.ListWebhooks(c.Request().Context(), workspaceID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Explicitly clear secrets from list response
	for i := range webhooks {
		webhooks[i].Secret = ""
	}

	return c.JSON(http.StatusOK, webhooks)
}

func (h *WebhookHandler) UpdateWebhook(c echo.Context) error {
	workspaceIDStr := c.Get("tenant_id").(string)
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid tenant_id"})
	}

	webhookIDStr := c.Param("id")
	webhookID, err := uuid.Parse(webhookIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid webhook id"})
	}

	var req webhook.UpdateWebhookPayload
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	updated, err := h.repo.UpdateWebhook(c.Request().Context(), workspaceID, webhookID, req.Active, req.EndpointURL, req.Events)
	if err != nil {
		if err == repository.ErrNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "webhook not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	updated.Secret = "" // hide secret
	return c.JSON(http.StatusOK, updated)
}

func (h *WebhookHandler) DeleteWebhook(c echo.Context) error {
	workspaceIDStr := c.Get("tenant_id").(string)
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid tenant_id"})
	}

	webhookIDStr := c.Param("id")
	webhookID, err := uuid.Parse(webhookIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid webhook id"})
	}

	err = h.repo.DeleteWebhook(c.Request().Context(), workspaceID, webhookID)
	if err != nil {
		if err == repository.ErrNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "webhook not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *WebhookHandler) ListDeliveries(c echo.Context) error {
	workspaceIDStr := c.Get("tenant_id").(string)
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid tenant_id"})
	}

	webhookIDStr := c.Param("id")
	webhookID, err := uuid.Parse(webhookIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid webhook id"})
	}

	// Verify the webhook belongs to the workspace
	_, err = h.repo.GetWebhook(c.Request().Context(), workspaceID, webhookID)
	if err != nil {
		if err == repository.ErrNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "webhook not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	deliveries, err := h.repo.ListDeliveries(c.Request().Context(), webhookID, 50)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, deliveries)
}
