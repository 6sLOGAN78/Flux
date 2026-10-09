package service_test

import (
	"testing"

	"github.com/6sLOGAN78/flux/internal/service"
	"github.com/stretchr/testify/require"
)

func TestWorkspacePermissionMatrix(t *testing.T) {
	for _, role := range []string{"owner", "admin", "member", "viewer", "unknown", "", "OWNER"} {
		t.Run(role, func(t *testing.T) {
			valid := role == "owner" || role == "admin" || role == "member" || role == "viewer"
			require.Equal(t, valid, service.Allows(role, service.CapabilityRead))
			require.Equal(t, valid && role != "viewer", service.Allows(role, service.CapabilityWrite))
			require.Equal(t, role == "owner" || role == "admin", service.Allows(role, service.CapabilityTeam))
			require.False(t, service.Allows(role, "unknown"))
		})
	}
}

func TestWorkspaceRoleChangeMatrix(t *testing.T) {
	roles := []string{"owner", "admin", "member", "viewer", "unknown"}
	for _, actor := range roles {
		for _, current := range roles {
			for _, proposed := range roles {
				valid := current != "unknown" && proposed != "unknown"
				want := valid && (actor == "owner" || actor == "admin" &&
					(current == "member" || current == "viewer") && (proposed == "member" || proposed == "viewer"))
				require.Equal(t, want, service.AllowsRoleChange(actor, current, proposed), "%s: %s -> %s", actor, current, proposed)
			}
		}
	}
}
