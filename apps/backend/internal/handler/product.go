package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/6sLOGAN78/flux/internal/errs"
	"github.com/6sLOGAN78/flux/internal/repository"
	"github.com/6sLOGAN78/flux/internal/service"
	"github.com/6sLOGAN78/flux/internal/transport"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const preconditionRequiredCode = "PRECONDITION_REQUIRED"
const jsonNull = "null"

// ProductHandler adapts verified control-plane identity to canonical transports.
type ProductHandler struct {
	auth        *service.AuthService
	identity    service.IdentityResolver
	workspace   *service.WorkspaceService
	links       *service.LinkService
	team        *service.TeamService
	invitations *service.InvitationService
}

// NewProductHandler injects the authentication boundary explicitly.
func NewProductHandler(auth *service.AuthService, identity service.IdentityResolver,
	workspace *service.WorkspaceService, links *service.LinkService, team *service.TeamService,
	invitations ...*service.InvitationService,
) *ProductHandler {
	h := &ProductHandler{auth: auth, identity: identity, workspace: workspace, links: links, team: team}
	if len(invitations) > 0 {
		h.invitations = invitations[0]
	}
	return h
}

func (h *ProductHandler) invitationScope(c echo.Context) (repository.Scope, error) {
	user, err := h.workspaceActor(c)
	if err != nil {
		return repository.Scope{}, err
	}
	id, err := uuid.Parse(c.Param("workspaceId"))
	if err != nil || id == uuid.Nil || id.String() != c.Param("workspaceId") {
		return repository.Scope{}, errs.NewNotFoundError("Workspace not found", false, nil)
	}
	if h.invitations == nil {
		return repository.Scope{}, workspaceUnavailable()
	}
	return repository.Scope{WorkspaceID: id, ActorID: user.ID}, nil
}

// CreateInvitation uses canonical DTOs and strict, bounded JSON input.
func (h *ProductHandler) CreateInvitation(c echo.Context) error {
	scope, err := h.invitationScope(c)
	if err != nil {
		return err
	}
	if c.Request().URL.RawQuery != "" {
		return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	var body *transport.TransportCreateInvitationRequest
	decoder := json.NewDecoder(c.Request().Body)
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&body); err != nil || body == nil {
		return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	var trailing any
	if err = decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	item, err := h.invitations.Create(c.Request().Context(), scope, body.Email, string(body.Role),
		c.Request().Header.Get("Idempotency-Key"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, transport.TransportInvitationResponse{Invitation: invitationTransport(item)})
}

// ListInvitations projects safe invitation fields with bounded UUID pagination.
func (h *ProductHandler) ListInvitations(c echo.Context) error {
	scope, err := h.invitationScope(c)
	if err != nil {
		return err
	}
	raw := c.Request().URL.RawQuery
	const maxQueryBytes = 256
	if len(raw) > maxQueryBytes {
		return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	query, err := url.ParseQuery(raw)
	if err != nil {
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
	items, next, err := h.invitations.List(c.Request().Context(), scope, after)
	if err != nil {
		return err
	}
	response := transport.TransportInvitationsResponse{
		Items: make([]transport.TransportInvitation, 0, len(items)), NextAfter: next,
	}
	for _, item := range items {
		response.Items = append(response.Items, invitationTransport(item))
	}
	return c.JSON(http.StatusOK, response)
}

func invitationTransport(item repository.Invitation) transport.TransportInvitation {
	return transport.TransportInvitation{Id: item.ID.String(), WorkspaceId: item.WorkspaceID.String(), Email: item.Email,
		Role: transport.TransportInvitationRole(item.Role), Status: transport.TransportInvitationStatus(item.Status),
		ExpiresAt: item.ExpiresAt.Format(time.RFC3339Nano), Version: item.Version}
}

// ResendInvitation creates a new immutable delivery identity after authorization.
func (h *ProductHandler) ResendInvitation(c echo.Context) error { return h.mutateInvitation(c, true) }

// RevokeInvitation atomically invalidates acceptance and old delivery leases.
func (h *ProductHandler) RevokeInvitation(c echo.Context) error { return h.mutateInvitation(c, false) }

func (h *ProductHandler) mutateInvitation(c echo.Context, resend bool) error {
	scope, err := h.invitationScope(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Param("invitationId"))
	if err != nil || id == uuid.Nil || id.String() != c.Param("invitationId") {
		return errs.NewNotFoundError("Invitation not found", false, nil)
	}
	if c.Request().URL.RawQuery != "" {
		return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	var body *transport.TransportInvitationMutationRequest
	var raw json.RawMessage
	decoder := json.NewDecoder(c.Request().Body)
	if err = decoder.Decode(&raw); err != nil {
		return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	fields := json.NewDecoder(bytes.NewReader(raw))
	fields.DisallowUnknownFields()
	if err = fields.Decode(&body); err != nil || body == nil {
		return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	var trailing any
	if err = decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	if body.Version == nil {
		var presence map[string]json.RawMessage
		if err = json.Unmarshal(raw, &presence); err != nil {
			return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
		}
		if _, present := presence["version"]; present {
			return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
		}
		return &errs.HTTPError{Code: preconditionRequiredCode, Message: "Reload invitations before trying again.",
			Status: http.StatusPreconditionRequired}
	}
	item, err := h.invitations.Mutate(c.Request().Context(), scope, id, *body.Version,
		c.Request().Header.Get("Idempotency-Key"), resend)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, transport.TransportInvitationResponse{Invitation: invitationTransport(item)})
}

// InspectInvitation never performs a membership effect.
func (h *ProductHandler) InspectInvitation(c echo.Context) error { return h.redeemInvitation(c, false) }

// AcceptInvitation is the only explicit invitation membership operation.
func (h *ProductHandler) AcceptInvitation(c echo.Context) error { return h.redeemInvitation(c, true) }

func (h *ProductHandler) redeemInvitation(c echo.Context, accept bool) error {
	user, err := h.workspaceActor(c)
	if err != nil {
		return err
	}
	if h.invitations == nil {
		return workspaceUnavailable()
	}
	actor, ok := c.Get("actor").(service.Actor)
	if !ok {
		return workspaceUnavailable()
	}
	if c.Request().URL.RawQuery != "" {
		return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	var body *transport.TransportInvitationTokenRequest
	decoder := json.NewDecoder(c.Request().Body)
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&body); err != nil || body == nil {
		return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	var trailing any
	if err = decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	review, workspace, err := h.invitations.Redeem(c.Request().Context(), user.ID, actor, h.auth, body.Token, accept)
	if err != nil {
		return err
	}
	if accept {
		return c.JSON(http.StatusOK, transport.TransportWorkspaceResponse{Workspace: workspaceTransport(workspace)})
	}
	response := transport.TransportInvitationInspectionResponse{}
	response.Invitation.ExpiresAt = review.ExpiresAt.Format(time.RFC3339Nano)
	response.Invitation.Role = transport.TransportInvitationInspectionRole(review.Role)
	response.Invitation.Status = transport.TransportInvitationInspectionStatus(review.Status)
	response.Invitation.WorkspaceId = review.WorkspaceID.String()
	response.Invitation.WorkspaceName = review.WorkspaceName
	return c.JSON(http.StatusOK, response)
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

// RemoveMember accepts the canonical empty JSON request and returns committed identity.
func (h *ProductHandler) RemoveMember(c echo.Context) error {
	user, err := h.workspaceActor(c)
	if err != nil {
		return err
	}
	workspace, err := uuid.Parse(c.Param("workspaceId"))
	if err != nil || workspace == uuid.Nil || workspace.String() != c.Param("workspaceId") {
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
	var body *transport.TransportRemoveMemberRequest
	decoder := json.NewDecoder(c.Request().Body)
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&body); err != nil || body == nil || len(*body) != 0 {
		return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	var trailing any
	if err = decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	result, err := h.team.Remove(c.Request().Context(), repository.Scope{WorkspaceID: workspace, ActorID: user.ID},
		target, c.Request().Header.Get("Idempotency-Key"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, transport.TransportMemberRemovalResponse{
		RemovedUserId: result.UserID.String(), WorkspaceId: result.WorkspaceID.String(), SelfRemoved: result.SelfRemoved,
	})
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
		if (name != "destination" && name != "title" && name != "customKey") || string(value) == jsonNull {
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

// TransitionLink accepts only the canonical lifecycle body and exact precondition.
func (h *ProductHandler) TransitionLink(c echo.Context) error {
	scope, err := h.linkScope(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Param("linkId"))
	if err != nil || id == uuid.Nil || id.String() != c.Param("linkId") {
		return errs.NewNotFoundError("Link not found", false, nil)
	}
	if c.Request().URL.RawQuery != "" {
		return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	var fields map[string]json.RawMessage
	decoder := json.NewDecoder(c.Request().Body)
	if err = decoder.Decode(&fields); err != nil || fields == nil {
		return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	for name, value := range fields {
		if (name != "action" && name != "expectedVersion") || string(value) == jsonNull {
			return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
		}
	}
	var trailing any
	if err = decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	payload, err := json.Marshal(fields)
	if err != nil {
		return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	var body transport.TransportLinkTransitionRequest
	if err = json.Unmarshal(payload, &body); err != nil {
		return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	if body.ExpectedVersion == nil {
		return &errs.HTTPError{Code: preconditionRequiredCode, Status: http.StatusPreconditionRequired,
			Message: "We couldn't verify the latest version. Reload this link before trying again."}
	}
	item, err := h.links.Transition(c.Request().Context(), scope, id, string(body.Action),
		*body.ExpectedVersion, c.Request().Header.Get("Idempotency-Key"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, transport.TransportLinkResponse{Link: linkTransport(item)})
}

// SuspensionLink adapts the canonical strict reason/version body for both methods.
func (h *ProductHandler) SuspensionLink(c echo.Context) error {
	scope, err := h.linkScope(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Param("linkId"))
	if err != nil || id == uuid.Nil || id.String() != c.Param("linkId") {
		return errs.NewNotFoundError("Link not found", false, nil)
	}
	if c.Request().URL.RawQuery != "" {
		return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	body, err := decodeSuspensionRequest(c.Request().Body)
	if err != nil {
		return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	if body.ExpectedVersion == nil {
		return &errs.HTTPError{Code: preconditionRequiredCode, Status: http.StatusPreconditionRequired,
			Message: "We couldn't verify the latest version. Reload this link before trying again."}
	}
	action := "suspend"
	if c.Request().Method == http.MethodDelete {
		action = "clear"
	}
	item, err := h.links.Suspension(c.Request().Context(), scope, id, action, body.Reason,
		*body.ExpectedVersion, c.Request().Header.Get("Idempotency-Key"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, transport.TransportLinkResponse{Link: linkTransport(item)})
}

func decodeSuspensionRequest(reader io.Reader) (transport.TransportLinkSuspensionRequest, error) {
	var body transport.TransportLinkSuspensionRequest
	raw, err := io.ReadAll(reader)
	if err != nil || !utf8.Valid(raw) {
		return body, errors.New("invalid suspension body")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return body, errors.New("invalid suspension object")
	}
	fields := make(map[string]json.RawMessage)
	for decoder.More() {
		token, err = decoder.Token()
		if err != nil {
			return body, err
		}
		name, ok := token.(string)
		if !ok || (name != "reason" && name != "expectedVersion") {
			return body, errors.New("invalid suspension field")
		}
		if _, duplicate := fields[name]; duplicate {
			return body, errors.New("duplicate suspension field")
		}
		var value json.RawMessage
		if err = decoder.Decode(&value); err != nil {
			return body, err
		}
		if string(value) == jsonNull {
			return body, errors.New("null suspension field")
		}
		fields[name] = value
	}
	if _, err = decoder.Token(); err != nil {
		return body, err
	}
	var trailing any
	if err = decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return body, errors.New("trailing suspension body")
	}
	payload, err := json.Marshal(fields)
	if err != nil {
		return body, err
	}
	err = json.Unmarshal(payload, &body)
	return body, err
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
