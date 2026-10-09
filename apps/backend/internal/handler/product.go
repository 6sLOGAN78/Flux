package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/6sLOGAN78/flux/internal/errs"
	"github.com/6sLOGAN78/flux/internal/repository"
	"github.com/6sLOGAN78/flux/internal/service"
	"github.com/6sLOGAN78/flux/internal/transport"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ProductHandler adapts verified control-plane identity to canonical transports.
type ProductHandler struct {
	auth      *service.AuthService
	identity  service.IdentityResolver
	workspace *service.WorkspaceService
}

// NewProductHandler injects the authentication boundary explicitly.
func NewProductHandler(auth *service.AuthService, identity service.IdentityResolver,
	workspace *service.WorkspaceService,
) *ProductHandler {
	return &ProductHandler{auth: auth, identity: identity, workspace: workspace}
}

func (h *ProductHandler) workspaceActor(c echo.Context) (repository.User, error) {
	actor, ok := c.Get("actor").(service.Actor)
	if !ok {
		return repository.User{}, errs.NewUnauthorizedError("Authentication required", false)
	}
	if h.identity == nil || h.workspace == nil {
		return repository.User{}, &errs.HTTPError{Code: "SERVICE_UNAVAILABLE",
			Message: "Workspace temporarily unavailable", Status: http.StatusServiceUnavailable}
	}
	return h.identity.Resolve(c.Request().Context(), actor)
}

func workspaceTransport(item repository.Workspace) transport.TransportWorkspace {
	return transport.TransportWorkspace{
		Id: item.ID.String(), Name: item.Name, Role: transport.TransportWorkspaceRole(item.Role)}
}

// CreateWorkspace accepts only the canonical JSON body and commits before returning.
func (h *ProductHandler) CreateWorkspace(c echo.Context) error {
	user, err := h.workspaceActor(c)
	if err != nil {
		return err
	}
	var body transport.TransportCreateWorkspaceRequest
	decoder := json.NewDecoder(c.Request().Body)
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&body); err != nil {
		return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	var trailing any
	if err = decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	item, err := h.workspace.Create(c.Request().Context(), user.ID, body.Name, c.Request().Header.Get("Idempotency-Key"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, transport.TransportWorkspaceResponse{Workspace: workspaceTransport(item)})
}

// ListWorkspaces returns only current server-authorized memberships.
func (h *ProductHandler) ListWorkspaces(c echo.Context) error {
	user, err := h.workspaceActor(c)
	if err != nil {
		return err
	}
	items, err := h.workspace.List(c.Request().Context(), user.ID)
	if err != nil {
		return err
	}
	response := transport.TransportWorkspacesResponse{Workspaces: make([]transport.TransportWorkspace, 0, len(items))}
	for _, item := range items {
		response.Workspaces = append(response.Workspaces, workspaceTransport(item))
	}
	return c.JSON(http.StatusOK, response)
}

// WorkspaceSummary authorizes fresh membership before exposing the tenant name.
func (h *ProductHandler) WorkspaceSummary(c echo.Context) error {
	user, err := h.workspaceActor(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Param("workspaceId"))
	if err != nil {
		return errs.NewNotFoundError("Workspace not found", false, nil)
	}
	item, err := h.workspace.Summary(c.Request().Context(), repository.Scope{WorkspaceID: id, ActorID: user.ID})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, transport.TransportWorkspaceResponse{Workspace: workspaceTransport(item)})
}

// Me returns the committed internal identity for an actively verified bearer.
func (h *ProductHandler) Me(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store")
	actor, ok := c.Get("actor").(service.Actor)
	if !ok {
		return errs.NewUnauthorizedError("Authentication required", false)
	}
	if h.identity == nil {
		return &errs.HTTPError{Code: "SERVICE_UNAVAILABLE", Message: "Authentication temporarily unavailable",
			Status: http.StatusServiceUnavailable}
	}
	user, err := h.identity.Resolve(c.Request().Context(), actor)
	if err != nil {
		return err
	}
	response := transport.TransportIdentityResponse{Authenticated: true}
	response.User.Id = user.ID.String()
	response.User.Email = user.Email
	return c.JSON(http.StatusOK, response)
}
