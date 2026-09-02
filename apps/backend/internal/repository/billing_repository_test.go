package repository_test

import (
	"context"
	"testing"
	"time"

	"flux/apps/backend/internal/database"
	"flux/apps/backend/internal/modules/billing"
	"flux/apps/backend/internal/repository"
	pkgtesting "flux/apps/backend/internal/testing"
	"github.com/rs/zerolog"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBillingRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping billing repository integration test in short mode")
	}

	ctx := context.Background()
	pgContainer, err := pkgtesting.SetupPostgresContainer(ctx)
	require.NoError(t, err)
	defer pgContainer.Terminate(context.Background())

	dbPool, err := database.InitDBPool(ctx, pgContainer.DSN)
	require.NoError(t, err)
	defer dbPool.Close()

	logger := zerolog.Nop()
	err = database.MigrateDSN(ctx, &logger, pgContainer.DSN)
	require.NoError(t, err)

	repo := repository.NewBillingRepository(dbPool)

	// Create workspace
	workspaceID := uuid.New()
	_, err = dbPool.Exec(ctx, `INSERT INTO workspaces (id, name) VALUES ($1, 'Test Workspace')`, workspaceID)
	require.NoError(t, err)

	// Insert subscription
	now := time.Now().UTC()
	sub := &billing.Subscription{
		WorkspaceID:          workspaceID,
		StripeCustomerID:     "cus_123",
		StripeSubscriptionID: "sub_123",
		PlanTier:             "pro",
		Status:               "active",
		CurrentPeriodStart:   now,
		CurrentPeriodEnd:     now.AddDate(0, 1, 0),
		CancelAtPeriodEnd:    false,
	}

	err = repo.UpsertSubscription(ctx, sub)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, sub.ID)

	// Fetch subscription
	fetched, err := repo.GetSubscriptionByWorkspace(ctx, workspaceID)
	require.NoError(t, err)
	assert.Equal(t, sub.StripeSubscriptionID, fetched.StripeSubscriptionID)
	assert.Equal(t, "pro", fetched.PlanTier)
	
	// Ensure workspace got updated with stripe_customer_id
	var customerID string
	err = dbPool.QueryRow(ctx, `SELECT stripe_customer_id FROM workspaces WHERE id = $1`, workspaceID).Scan(&customerID)
	require.NoError(t, err)
	assert.Equal(t, "cus_123", customerID)

	// Upsert subscription again
	sub.PlanTier = "business"
	err = repo.UpsertSubscription(ctx, sub)
	require.NoError(t, err)
	
	// Verify update
	fetched2, err := repo.GetSubscriptionByWorkspace(ctx, workspaceID)
	require.NoError(t, err)
	assert.Equal(t, "business", fetched2.PlanTier)

	// Test uniqueness (Stripe subscription ID uniqueness)
	workspace2 := uuid.New()
	_, err = dbPool.Exec(ctx, `INSERT INTO workspaces (id, name) VALUES ($1, 'Test Workspace 2')`, workspace2)
	require.NoError(t, err)

	sub2 := &billing.Subscription{
		WorkspaceID:          workspace2,
		StripeCustomerID:     "cus_456",
		StripeSubscriptionID: "sub_123", // Duplicate
		PlanTier:             "pro",
		Status:               "active",
		CurrentPeriodStart:   now,
		CurrentPeriodEnd:     now.AddDate(0, 1, 0),
	}
	err = repo.UpsertSubscription(ctx, sub2)
	require.Error(t, err)
}
