package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

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
	links     *service.LinkService
	team      *service.TeamService
}

// NewProductHandler injects the authentication boundary explicitly.
func NewProductHandler(auth *service.AuthService, identity service.IdentityResolver,
	workspace *service.WorkspaceService, links *service.LinkService, team *service.TeamService,
) *ProductHandler {
	return &ProductHandler{auth: auth, identity: identity, workspace: workspace, links: links, team: team}
}

// ListMembers exposes only current workspace-bound identities and closed roles.
func (h *ProductHandler) ListMembers(c echo.Context) error {
	user, err := h.workspaceActor(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Param("workspaceId"))
	if err != nil || id.String() != c.Param("workspaceId") {
		return errs.NewNotFoundError("Workspace not found", false, nil)
	}
	if h.team == nil {
		return workspaceUnavailable()
	}
	query, err := url.ParseQuery(c.Request().URL.RawQuery)
	if err != nil || len(c.Request().URL.RawQuery) > 256 {
		return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	var after uuid.UUID
	for name, values := range query {
		if name != "after" || len(values) != 1 {
			return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
		}
		after, err = uuid.Parse(values[0])
		if err != nil || after == uuid.Nil || after.String() != values[0] {
			return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
		}
	}
	items, next, err := h.team.List(c.Request().Context(), repository.Scope{WorkspaceID: id, ActorID: user.ID}, after)
	if err != nil {
		return err
	}
	response := transport.TransportMembersResponse{
		Items: make([]transport.TransportTeamMember, 0, len(items)), NextAfter: next,
	}
	for _, item := range items {
		response.Items = append(response.Items, transport.TransportTeamMember{Id: item.ID.String(),
			WorkspaceId: item.WorkspaceID.String(), Email: item.Email, Role: transport.TransportTeamMemberRole(item.Role)})
	}
	return c.JSON(http.StatusOK, response)
}

func workspaceUnavailable() error {
	return &errs.HTTPError{Code: serviceUnavailableCode,
		Message: "Workspace temporarily unavailable", Status: http.StatusServiceUnavailable}
}

// ChangeMemberRole accepts only the canonical role DTO and returns committed state.
func (h *ProductHandler) ChangeMemberRole(c echo.Context) error {
	user, err := h.workspaceActor(c)
	if err != nil {
		return err
	}
	workspace, err := uuid.Parse(c.Param("workspaceId"))
	if err != nil || workspace.String() != c.Param("workspaceId") {
		return errs.NewNotFoundError("Member not found", false, nil)
	}
	target, err := uuid.Parse(c.Param("memberId"))
	if err != nil || target == uuid.Nil || target.String() != c.Param("memberId") {
		return errs.NewNotFoundError("Member not found", false, nil)
	}
	if h.team == nil {
		return workspaceUnavailable()
	}
	if c.Request().URL.RawQuery != "" {
		return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	var body transport.TransportChangeMemberRoleRequest
	decoder := json.NewDecoder(c.Request().Body)
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&body); err != nil {
		return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	var trailing any
	if err = decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	item, err := h.team.ChangeRole(c.Request().Context(), repository.Scope{WorkspaceID: workspace, ActorID: user.ID},
		target, string(body.Role), c.Request().Header.Get("Idempotency-Key"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, transport.TransportMemberResponse{
		ActorRole: transport.TransportMemberResponseActorRole(item.ActorRole),
		Member: transport.TransportTeamMember{Id: item.ID.String(), WorkspaceId: item.WorkspaceID.String(),
			Email: item.Email, Role: transport.TransportTeamMemberRole(item.Role)},
	})
}

const serviceUnavailableCode = "SERVICE_UNAVAILABLE"

func (h *ProductHandler) workspaceActor(c echo.Context) (repository.User, error) {
	actor, ok := c.Get("actor").(service.Actor)
	if !ok {
		return repository.User{}, errs.NewUnauthorizedError("Authentication required", false)
	}
	if h.identity == nil || h.workspace == nil {
		return repository.User{}, &errs.HTTPError{Code: serviceUnavailableCode,
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
	response := transport.TransportWorkspaceResponse{Workspace: workspaceTransport(item)}
	if h.links != nil {
		host := h.links.ManagedHost()
		response.ManagedHost = &host
	}
	return c.JSON(http.StatusOK, response)
}

func (h *ProductHandler) linkScope(c echo.Context) (repository.Scope, error) {
	user, err := h.workspaceActor(c)
	if err != nil {
		return repository.Scope{}, err
	}
	id, err := uuid.Parse(c.Param("workspaceId"))
	if err != nil || id.String() != c.Param("workspaceId") {
		return repository.Scope{}, errs.NewNotFoundError("Link not found", false, nil)
	}
	if h.links == nil {
		return repository.Scope{}, &errs.HTTPError{Code: errs.MakeUpperCaseWithUnderscores(
			http.StatusText(http.StatusServiceUnavailable)), Message: "Links temporarily unavailable",
			Status: http.StatusServiceUnavailable}
	}
	return repository.Scope{WorkspaceID: id, ActorID: user.ID}, nil
}

// CreateLink accepts only canonical DTO fields; service policy remains authority.
func (h *ProductHandler) CreateLink(c echo.Context) error {
	scope, err := h.linkScope(c)
	if err != nil {
		return err
	}
	var body transport.TransportCreateLinkRequest
	decoder := json.NewDecoder(c.Request().Body)
	decoder.DisallowUnknownFields()
	var fields map[string]json.RawMessage
	if err = decoder.Decode(&fields); err != nil || fields == nil {
		return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	for name, value := range fields {
		if (name != "destination" && name != "title" && name != "customKey") || string(value) == "null" {
			return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
		}
	}
	payload, err := json.Marshal(fields)
	if err != nil {
		return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	if err = json.Unmarshal(payload, &body); err != nil {
		return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	var trailing any
	if err = decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	title := ""
	if body.Title != nil {
		title = *body.Title
	}
	customKey := ""
	if body.CustomKey != nil {
		customKey = *body.CustomKey
		if customKey == "" {
			return errs.NewBadRequestError("Enter a valid custom short key.", false, nil, nil, nil)
		}
	}
	item, err := h.links.Create(c.Request().Context(), scope, body.Destination, title,
		c.Request().Header.Get("Idempotency-Key"), customKey)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, transport.TransportLinkResponse{Link: linkTransport(item)})
}

// LinkDetail reauthorizes before exposing any scoped protected values.
func (h *ProductHandler) LinkDetail(c echo.Context) error {
	scope, err := h.linkScope(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Param("linkId"))
	if err != nil || id.String() != c.Param("linkId") {
		return errs.NewNotFoundError("Link not found", false, nil)
	}
	item, err := h.links.Detail(c.Request().Context(), scope, id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, transport.TransportLinkResponse{Link: linkTransport(item)})
}

func linkTransport(item repository.Link) transport.TransportLink {
	result := transport.TransportLink{Id: item.ID.String(), WorkspaceId: item.WorkspaceID.String(),
		ShortUrl: "https://" + item.Host + "/" + item.Key, Destination: item.Destination, Title: item.Title,
		Lifecycle: transport.TransportLinkLifecycle(item.Lifecycle), Version: strconv.FormatInt(item.Version, 10),
		CreatedAt: item.CreatedAt.Format(time.RFC3339Nano), UpdatedAt: item.UpdatedAt.Format(time.RFC3339Nano)}
	result.Creator.Id = item.CreatorID.String()
	result.Creator.Email = item.CreatorEmail
	if item.SuspendedAt != nil && item.SuspendedBy != nil && item.SuspensionReason != nil {
		result.Suspension = &transport.TransportLinkSuspension{ActorId: item.SuspendedBy.String(),
			At: item.SuspendedAt.Format(time.RFC3339Nano), Reason: *item.SuspensionReason}
	}
	return result
}

// Me returns the committed internal identity for an actively verified bearer.
func (h *ProductHandler) Me(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store")
	actor, ok := c.Get("actor").(service.Actor)
	if !ok {
		return errs.NewUnauthorizedError("Authentication required", false)
	}
	if h.identity == nil || h.workspace == nil {
		return &errs.HTTPError{Code: serviceUnavailableCode, Message: "Authentication temporarily unavailable",
			Status: http.StatusServiceUnavailable}
	}
	user, err := h.identity.Resolve(c.Request().Context(), actor)
	if err != nil {
		return err
	}
	response := transport.TransportIdentityResponse{Authenticated: true}
	response.User.Id = user.ID.String()
	response.User.Email = user.Email
	items, last, err := h.workspace.Bootstrap(c.Request().Context(), user.ID)
	if err != nil {
		return err
	}
	response.Workspaces = make([]transport.TransportWorkspace, 0, len(items))
	for _, item := range items {
		response.Workspaces = append(response.Workspaces, workspaceTransport(item))
	}
	if last != nil {
		selected := workspaceTransport(*last)
		response.LastWorkspace = &selected
	}
	return c.JSON(http.StatusOK, response)
}

// SelectWorkspace validates the canonical selection body and commits fresh authority.
func (h *ProductHandler) SelectWorkspace(c echo.Context) error {
	user, err := h.workspaceActor(c)
	if err != nil {
		return err
	}
	var body transport.TransportWorkspacePreferenceRequest
	decoder := json.NewDecoder(c.Request().Body)
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&body); err != nil {
		return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	var trailing any
	if err = decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	id, err := uuid.Parse(body.WorkspaceId)
	if err != nil || id == uuid.Nil {
		return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	item, err := h.workspace.Select(c.Request().Context(), repository.Scope{WorkspaceID: id, ActorID: user.ID})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, transport.TransportWorkspaceResponse{Workspace: workspaceTransport(item)})
}

// ListLinks accepts bounded literal search, closed lifecycle and signed pagination.
func (h *ProductHandler) ListLinks(c echo.Context) error {
	scope, err := h.linkScope(c)
	if err != nil {
		return err
	}
	limit, cursor, filters, err := linkListQuery(c.Request().URL.RawQuery)
	if err != nil {
		return err
	}
	items, next, err := h.links.List(c.Request().Context(), scope, limit, cursor, filters.Search, filters.State)
	if err != nil {
		return err
	}
	response := transport.TransportLinksResponse{Items: make([]transport.TransportLink, 0, len(items)), NextCursor: next}
	for _, item := range items {
		response.Items = append(response.Items, linkTransport(item))
	}
	return c.JSON(http.StatusOK, response)
}

func linkListQuery(raw string) (int, string, service.CursorFilters, error) {
	const maxQueryBytes = 8192
	invalid := func() (int, string, service.CursorFilters, error) {
		return 0, "", service.CursorFilters{}, errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	invalidCursor := func() (int, string, service.CursorFilters, error) {
		return 0, "", service.CursorFilters{}, &errs.HTTPError{Code: "CURSOR_INVALID", Status: http.StatusBadRequest,
			Message: "This page is no longer available. Return to the first page."}
	}
	if len(raw) > maxQueryBytes {
		return invalidCursor()
	}
	query, err := url.ParseQuery(raw)
	if err != nil {
		return invalid()
	}
	for name, values := range query {
		if (name != "limit" && name != "cursor" && name != "search" && name != "state") || len(values) != 1 {
			return invalid()
		}
	}
	filters := service.CursorFilters{Search: query.Get("search"), State: query.Get("state")}
	if query.Has("state") && filters.State == "" {
		return invalid()
	}
	cursor := query.Get("cursor")
	if query.Has("cursor") && (cursor == "" || len(cursor) > 2048) {
		return invalidCursor()
	}
	limit, ok := linkListLimit(query)
	if !ok {
		return invalid()
	}
	return limit, cursor, filters, nil
}

func linkListLimit(query url.Values) (int, bool) {
	const defaultLimit = 25
	if !query.Has("limit") {
		return defaultLimit, true
	}
	value := query.Get("limit")
	if len(value) > 3 || value == "" {
		return 0, false
	}
	for _, digit := range value {
		if digit < '0' || digit > '9' {
			return 0, false
		}
	}
	limit, err := strconv.Atoi(value)
	return limit, err == nil && limit >= 1 && limit <= 100
}
