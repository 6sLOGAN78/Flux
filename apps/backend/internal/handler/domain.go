package handler

import (
	"errors"
	"net/http"

	"flux/apps/backend/internal/errs"
	"flux/apps/backend/internal/model/domain"
	"flux/apps/backend/internal/repository"
	"flux/apps/backend/internal/service"

	"github.com/labstack/echo/v4"
)

type DomainHandler struct {
	svc *service.DomainService
}

func NewDomainHandler(svc *service.DomainService) *DomainHandler {
	return &DomainHandler{svc: svc}
}

type CreateDomainRequest struct {
	Hostname string `json:"hostname"`
}

type CreateDomainResponse struct {
	ID                string `json:"id"`
	Hostname          string `json:"hostname"`
	Status            string `json:"status"`
	VerificationToken string `json:"verification_token"` // Only exposed at creation
}

func (h *DomainHandler) getTenantID(c echo.Context) string {
	tenantID, _ := c.Get("tenant_id").(string)
	return tenantID
}

func (h *DomainHandler) CreateDomain(c echo.Context) error {
	tenantID := h.getTenantID(c)
	if tenantID == "" {
		code := "UNAUTHORIZED"
		return &errs.HTTPError{StatusCode: http.StatusUnauthorized, Message: "missing tenant context", Code: &code}
	}

	var req CreateDomainRequest
	if err := c.Bind(&req); err != nil {
		return errs.NewBadRequestError("invalid request body", false, nil, nil, err)
	}

	d, err := h.svc.CreateDomain(c.Request().Context(), tenantID, req.Hostname)
	if err != nil {
		return err
	}

	res := CreateDomainResponse{
		ID:                d.ID,
		Hostname:          d.Hostname,
		Status:            d.Status,
		VerificationToken: d.VerificationToken,
	}

	return c.JSON(http.StatusCreated, res)
}

func (h *DomainHandler) GetDomains(c echo.Context) error {
	tenantID := h.getTenantID(c)
	if tenantID == "" {
		code := "UNAUTHORIZED"
		return &errs.HTTPError{StatusCode: http.StatusUnauthorized, Message: "missing tenant context", Code: &code}
	}

	domains, err := h.svc.GetDomains(c.Request().Context(), tenantID)
	if err != nil {
		return err
	}
	
	if domains == nil {
		domains = make([]domain.CustomDomain, 0) // return [] instead of null
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"data": domains})
}

func (h *DomainHandler) GetDomain(c echo.Context) error {
	tenantID := h.getTenantID(c)
	if tenantID == "" {
		code := "UNAUTHORIZED"
		return &errs.HTTPError{StatusCode: http.StatusUnauthorized, Message: "missing tenant context", Code: &code}
	}

	id := c.Param("id")
	if id == "" {
		return errs.NewBadRequestError("domain id is required", false, nil, nil, nil)
	}

	d, err := h.svc.GetDomainByID(c.Request().Context(), tenantID, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return errs.NewNotFoundError("domain not found or inaccessible", false, nil)
		}
		return err
	}

	return c.JSON(http.StatusOK, d)
}

func (h *DomainHandler) DeleteDomain(c echo.Context) error {
	tenantID := h.getTenantID(c)
	if tenantID == "" {
		code := "UNAUTHORIZED"
		return &errs.HTTPError{StatusCode: http.StatusUnauthorized, Message: "missing tenant context", Code: &code}
	}

	id := c.Param("id")
	if id == "" {
		return errs.NewBadRequestError("domain id is required", false, nil, nil, nil)
	}

	err := h.svc.DeleteDomain(c.Request().Context(), tenantID, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return errs.NewNotFoundError("domain not found or inaccessible", false, nil)
		}
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
