//nolint:testpackage // Verify the private closed state machine without exporting a mutation bypass.
package service

import (
	"testing"

	"github.com/6sLOGAN78/flux/internal/errs"
	"github.com/stretchr/testify/require"
)

func TestLifecycleClosedStateMachine(t *testing.T) {
	t.Parallel()
	allowed := map[string]string{
		"active/disable": "disabled", "disabled/enable": "active",
		"active/archive": "archived", "disabled/archive": "archived",
		"archived/restore": "disabled", "deleted/restore": "disabled",
		"active/delete": "deleted", "disabled/delete": "deleted", "archived/delete": "deleted",
	}
	for _, state := range []string{"active", "disabled", "archived", "deleted"} {
		for _, action := range []string{"enable", "disable", "archive", "restore", "delete"} {
			for _, suspended := range []bool{false, true} {
				got, err := nextLinkLifecycle(state, action, suspended)
				want := allowed[state+"/"+action]
				if want == "" || (suspended && action == "enable") {
					var typed *errs.HTTPError
					require.ErrorAs(t, err, &typed)
					require.Equal(t, 409, typed.Status)
				} else {
					require.NoError(t, err)
					require.Equal(t, want, got)
				}
			}
		}
	}
	require.False(t, validLinkAction("edit"))
}
