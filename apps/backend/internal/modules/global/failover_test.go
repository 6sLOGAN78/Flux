package global_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"flux/apps/backend/internal/modules/global"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCircuitBreaker_TrippingAndRecovery(t *testing.T) {
	cb := global.NewCircuitBreaker(2, 100*time.Millisecond)

	assert.Equal(t, global.StateClosed, cb.State())

	// First failure - still closed
	err := cb.Execute(func() error { return errors.New("db error") })
	assert.Error(t, err)
	assert.Equal(t, global.StateClosed, cb.State())

	// Second failure - trips to Open state
	err = cb.Execute(func() error { return errors.New("db error") })
	assert.Error(t, err)
	assert.Equal(t, global.StateOpen, cb.State())

	// Blocked while open
	err = cb.Execute(func() error { return errors.New("should not execute") })
	assert.True(t, errors.Is(err, global.ErrCircuitOpen))

	// Wait for timeout to transition to HalfOpen
	time.Sleep(150 * time.Millisecond)

	// Successful execution closes circuit
	err = cb.Execute(func() error { return nil })
	assert.NoError(t, err)
	assert.Equal(t, global.StateClosed, cb.State())
}

func TestFailoverManager_AutomatedFailover(t *testing.T) {
	fm := global.NewFailoverManager("us-east", []string{"eu-west", "ap-east"})

	ctx := context.Background()
	status, err := fm.GetClusterState(ctx)
	require.NoError(t, err)
	assert.Equal(t, "us-east", status.ActiveRegion)
	assert.False(t, status.IsFailedOver)

	// Simulate failures in primary active region
	var result *global.FailoverResult
	for i := 0; i < 3; i++ {
		res, err := fm.RecordRegionFailure("us-east")
		require.NoError(t, err)
		result = res
	}

	require.NotNil(t, result)
	assert.True(t, result.FailoverTriggered)
	assert.Equal(t, "us-east", result.PreviousRegion)
	assert.Equal(t, "eu-west", result.NewActiveRegion)

	status, err = fm.GetClusterState(ctx)
	require.NoError(t, err)
	assert.Equal(t, "eu-west", status.ActiveRegion)
	assert.True(t, status.IsFailedOver)
}
