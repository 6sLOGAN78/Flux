//nolint:testpackage // Verify that the private error boundary never exposes underlying store diagnostics.
package service

import (
	"context"
	"errors"
	"testing"

	"github.com/6sLOGAN78/flux/internal/errs"
	"github.com/6sLOGAN78/flux/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceSafeFailures(t *testing.T) {
	for _, tc := range []struct {
		cause  error
		status int
	}{
		{errors.New("PRIVATE-DATABASE-MARKER"), 503},
		{pgx.ErrNoRows, 404},
		{repository.ErrWorkspaceConflict, 409},
	} {
		var response *errs.HTTPError
		require.ErrorAs(t, workspaceFailure(tc.cause), &response)
		require.Equal(t, tc.status, response.Status)
		require.NotContains(t, response.Error(), "PRIVATE-DATABASE-MARKER")
	}
}

func TestWorkspaceInputValidation(t *testing.T) {
	svc := NewWorkspaceService(repository.NewWorkspaceRepository(nil))
	for _, tc := range []struct{ name, key string }{
		{" ", "workspace-valid-key"}, {"Growth", "short"}, {"Growth", "invalid/key-with-slash"},
	} {
		_, err := svc.Create(context.Background(), uuid.New(), tc.name, tc.key)
		var response *errs.HTTPError
		require.ErrorAs(t, err, &response)
		require.Equal(t, 400, response.Status)
	}
}
