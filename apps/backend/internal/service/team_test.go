//nolint:testpackage // Verify the closed private role policy and pre-store validation directly.
package service

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/6sLOGAN78/flux/internal/errs"
	"github.com/6sLOGAN78/flux/internal/repository"
	"github.com/google/uuid"
)

func TestTeamRolePolicy(t *testing.T) {
	for _, actor := range []string{"owner", "admin", "member", "viewer", "org:owner", ""} {
		for _, target := range []string{"owner", "admin", "member", "viewer", "unknown"} {
			for _, grant := range []string{"owner", "admin", "member", "viewer", "unknown"} {
				want := target != "unknown" && grant != "unknown" && (actor == "owner" ||
					(actor == "admin" && (target == "member" || target == "viewer") && (grant == "member" || grant == "viewer")))
				if AllowsRoleChange(actor, target, grant) != want {
					t.Fatalf("incorrect role policy for %s/%s/%s", actor, target, grant)
				}
			}
		}
	}
}

func TestTeamRoleInput(t *testing.T) {
	var svc *TeamService
	for _, input := range []struct {
		role, key string
		target    uuid.UUID
	}{
		{"OWNER", "valid-retry-key-0001", uuid.New()},
		{"member", "short", uuid.New()},
		{"member", strings.Repeat("k", 129), uuid.New()},
		{"member", "invalid retry key", uuid.New()},
		{"member", "valid-retry-key-0001", uuid.Nil},
	} {
		_, err := svc.ChangeRole(context.Background(), repository.Scope{}, input.target, input.role, input.key)
		var failure *errs.HTTPError
		if !errors.As(err, &failure) || failure.Status != http.StatusBadRequest {
			t.Fatal("invalid external input reached store")
		}
	}
	for _, role := range []string{"owner", "admin", "member", "viewer"} {
		_, err := svc.ChangeRole(context.Background(), repository.Scope{}, uuid.New(), role, "valid-retry-key-0001")
		var failure *errs.HTTPError
		if !errors.As(err, &failure) || failure.Status != http.StatusServiceUnavailable {
			t.Fatal("valid role did not reach unavailable store boundary")
		}
	}
}
