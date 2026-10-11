package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/6sLOGAN78/flux/internal/errs"
	"github.com/6sLOGAN78/flux/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestInvitationRedeemFailsClosedWithoutTrustedProfile(t *testing.T) {
	var invitations *service.InvitationService
	for _, token := range []string{"", "private-input", strings.Repeat("A", 42), strings.Repeat("A", 44), strings.Repeat("_", 43)} {
		review, workspace, err := invitations.Redeem(context.Background(), uuid.New(), service.Actor{}, nil, token, true)
		var response *errs.HTTPError
		require.True(t, errors.As(err, &response))
		require.Equal(t, 400, response.Status)
		require.Empty(t, review)
		require.Empty(t, workspace)
		require.Equal(t, "Invalid request", response.Message)
	}
	for _, accept := range []bool{false, true} {
		review, workspace, err := invitations.Redeem(context.Background(), uuid.New(), service.Actor{}, nil, strings.Repeat("A", 43), accept)
		var response *errs.HTTPError
		require.True(t, errors.As(err, &response))
		require.Equal(t, 503, response.Status)
		require.Empty(t, review)
		require.Empty(t, workspace)
	}
}
