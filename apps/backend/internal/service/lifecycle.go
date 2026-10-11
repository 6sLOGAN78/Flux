package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"math"
	"net/http"
	"regexp"
	"strconv"

	"github.com/6sLOGAN78/flux/internal/errs"
	"github.com/6sLOGAN78/flux/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	linkActive   = "active"
	linkDisabled = "disabled"
	linkArchived = "archived"
	linkDeleted  = "deleted"
)

// Transition validates an exact SDK-safe version and freshly authorizes every
// effect and replay within the workspace-first transaction.
func (s *LinkService) Transition(ctx context.Context, scope repository.Scope, id uuid.UUID,
	action, expectedVersion, key string,
) (repository.Link, error) {
	version, err := strconv.ParseInt(expectedVersion, 10, 64)
	versionOK, _ := regexp.MatchString(`^[1-9][0-9]{0,18}$`, expectedVersion)
	keyOK, _ := regexp.MatchString(`^[A-Za-z0-9_-]{16,128}$`, key)
	if err != nil || !versionOK || version < 1 || !keyOK || !validLinkAction(action) {
		return repository.Link{}, errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	payload, err := json.Marshal(struct {
		Action  string    `json:"action"`
		Version string    `json:"expectedVersion"`
		ID      uuid.UUID `json:"id"`
	}{action, expectedVersion, id})
	if err != nil {
		return repository.Link{}, linkFailure(err)
	}
	hash := sha256.Sum256(payload)
	authorize := func(ctx context.Context, tx pgx.Tx, scope repository.Scope) error {
		if _, lockErr := s.workspace.store.LockScope(ctx, tx, scope, false); lockErr != nil {
			return workspaceFailure(lockErr)
		}
		var role string
		if lockErr := tx.QueryRow(ctx, "SELECT role FROM memberships WHERE workspace_id=$1 AND user_id=$2 FOR UPDATE",
			scope.WorkspaceID, scope.ActorID).Scan(&role); lockErr != nil {
			return workspaceFailure(lockErr)
		}
		if !Allows(role, CapabilityWrite) {
			return errs.NewForbiddenError("You do not have permission for this action.", false)
		}
		return nil
	}
	apply := func(link repository.Link) (string, error) {
		if link.Version != version || version == math.MaxInt64 {
			return "", lifecycleConflict("VERSION_CONFLICT",
				"This link changed since you opened it. Review the latest details before trying again.")
		}
		return nextLinkLifecycle(link.Lifecycle, action, link.SuspendedAt != nil)
	}
	ctx, cancel := context.WithTimeout(ctx, workspaceTimeout)
	defer cancel()
	result, err := s.store.Transition(ctx, scope, id, version, action, key, hash[:], authorize, apply)
	return result, linkFailure(err)
}

func validLinkAction(action string) bool {
	switch action {
	case "enable", "disable", "archive", "restore", "delete":
		return true
	}
	return false
}

func nextLinkLifecycle(state, action string, suspended bool) (string, error) {
	switch {
	case state == linkActive && action == "disable":
		return linkDisabled, nil
	case state == linkDisabled && action == "enable":
		if suspended {
			return "", lifecycleConflict("SUSPENDED", "An owner or admin must clear suspension first.")
		}
		return linkActive, nil
	case (state == linkActive || state == linkDisabled) && action == "archive":
		return linkArchived, nil
	case (state == linkArchived || state == linkDeleted) && action == "restore":
		return linkDisabled, nil
	case state != linkDeleted && action == "delete":
		return linkDeleted, nil
	default:
		return "", lifecycleConflict("INVALID_TRANSITION", "This lifecycle action is unavailable. Review the latest details.")
	}
}

func lifecycleConflict(code, message string) error {
	return &errs.HTTPError{Code: code, Message: message, Status: http.StatusConflict}
}
