package handler

import (
	"fmt"
	"net/http"

	"flux/apps/backend/internal/errs"
	"flux/apps/backend/internal/model/link"
	"flux/apps/backend/internal/service"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type LinksHandler struct {
	svc *service.LinkService
}

func NewLinksHandler(svc *service.LinkService) *LinksHandler {
	return &LinksHandler{svc: svc}
}

func (h *LinksHandler) getTenantID(c echo.Context) *uuid.UUID {
	tID, ok := c.Get("tenant_id").(uuid.UUID)
	if !ok {
		return nil
	}
	return &tID
}

func (h *LinksHandler) CreateLink(c echo.Context) error {
	var payload link.CreateLinkPayload
	if err := c.Bind(&payload); err != nil {
		fmt.Printf("Error in handler: %+v\n", err); return err
	}
	if err := payload.Validate(); err != nil {
		fmt.Printf("Error in handler: %+v\n", err); return errs.NewBadRequestError("validation failed", true, nil, nil, err)
	}

	tenantID := h.getTenantID(c)
	res, err := h.svc.CreateLink(c.Request().Context(), tenantID, &payload)
	if err != nil {
		fmt.Printf("Error in handler: %+v\n", err); return err
	}

	return c.JSON(http.StatusCreated, res)
}

func (h *LinksHandler) GetLinks(c echo.Context) error {
	var query link.GetLinksQuery
	if err := c.Bind(&query); err != nil {
		fmt.Printf("Error in handler: %+v\n", err); return err
	}
	if err := query.Validate(); err != nil {
		fmt.Printf("Error in handler: %+v\n", err); return errs.NewBadRequestError("validation failed", true, nil, nil, err)
	}

	tenantID := h.getTenantID(c)
	res, err := h.svc.GetLinks(c.Request().Context(), tenantID, &query)
	if err != nil {
		fmt.Printf("Error in handler: %+v\n", err); return err
	}

	return c.JSON(http.StatusOK, res)
}

func (h *LinksHandler) GetLinkByID(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		fmt.Printf("Error in handler: %+v\n", err); return errs.NewBadRequestError("invalid id format", false, nil, nil, nil)
	}

	tenantID := h.getTenantID(c)
	res, err := h.svc.GetLinkByID(c.Request().Context(), tenantID, id)
	if err != nil {
		fmt.Printf("Error in handler: %+v\n", err); return err
	}

	return c.JSON(http.StatusOK, res)
}

func (h *LinksHandler) UpdateLink(c echo.Context) error {
	var payload link.UpdateLinkPayload
	if err := c.Bind(&payload); err != nil {
		fmt.Printf("Error in handler: %+v\n", err); return err
	}
	
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		fmt.Printf("Error in handler: %+v\n", err); return errs.NewBadRequestError("invalid id format", false, nil, nil, nil)
	}
	payload.ID = id

	if err := payload.Validate(); err != nil {
		fmt.Printf("Error in handler: %+v\n", err); return errs.NewBadRequestError("validation failed", true, nil, nil, err)
	}

	res, err := h.svc.UpdateLink(c.Request().Context(), h.getTenantID(c), &payload)
	if err != nil {
		fmt.Printf("Error in handler: %+v\n", err); return err
	}

	return c.JSON(http.StatusOK, res)
}

func (h *LinksHandler) DeleteLink(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		fmt.Printf("Error in handler: %+v\n", err); return errs.NewBadRequestError("invalid id format", false, nil, nil, nil)
	}

	if err := h.svc.DeleteLink(c.Request().Context(), h.getTenantID(c), id); err != nil {
		fmt.Printf("Error in handler: %+v\n", err); return err
	}

	return c.NoContent(http.StatusNoContent)
}
