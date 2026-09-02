package service_test

import (
	"context"
	"testing"
	"time"

	"flux/apps/backend/internal/database"
	"flux/apps/backend/internal/model/link"
	"flux/apps/backend/internal/repository"
	"flux/apps/backend/internal/service"
	pkgtesting "flux/apps/backend/internal/testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinkService_Quota(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	pgContainer, err := pkgtesting.SetupPostgresContainer(ctx)
	require.NoError(t, err)
	defer pgContainer.Terminate(ctx)

	dbPool, err := database.InitDBPool(ctx, pgContainer.DSN)
	require.NoError(t, err)
	defer dbPool.Close()

	logger := zerolog.Nop()
	err = database.MigrateDSN(ctx, &logger, pgContainer.DSN)
	require.NoError(t, err)

	linkRepo := repository.NewLinkRepository(dbPool)
	campRepo := repository.NewCampaignRepository(dbPool)
	billingRepo := repository.NewBillingRepository(dbPool)
	
	linkSvc := service.NewLinkService(linkRepo, nil, campRepo, billingRepo)

	// Create workspace A (Free)
	workspaceA := uuid.New()
	_, err = dbPool.Exec(ctx, `INSERT INTO workspaces (id, name, stripe_customer_id) VALUES ($1, 'Workspace A', 'cus_free')`, workspaceA)
	require.NoError(t, err)

	// Workspace A doesn't have a subscription explicitly, defaults to Free -> Max Links: 1000
	payload := &link.CreateLinkPayload{DestinationURL: "https://example.com"}
	
	// Create link under quota
	_, err = linkSvc.CreateLink(ctx, &workspaceA, payload)
	require.NoError(t, err)

	// Since 1000 is too much to insert manually for testing, let's create Workspace B with a Pro subscription and limit it to 2 via a trick?
	// Wait, we can't change the limit easily as it is hardcoded in `billing.GetTierLimits(planTier)`.
	// Let's test the concurrency instead. Wait, inserting 1000 links is quick enough for a test. Let's do it!
	
	t.Run("Free Tier Exhaustion", func(t *testing.T) {
		// Fill it up to 999 more links
		for i := 0; i < 999; i++ {
			_, err := linkSvc.CreateLink(ctx, &workspaceA, payload)
			require.NoError(t, err, "failed at link %d", i)
		}
		
		// Attempting 1001st link should fail with QuotaExceeded
		_, err = linkSvc.CreateLink(ctx, &workspaceA, payload)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Link limit reached")
	})

	t.Run("Paid Tier Expands Quota", func(t *testing.T) {
		// Upgrade Workspace A to Pro
		_, err = dbPool.Exec(ctx, `
			INSERT INTO subscriptions (
				workspace_id, stripe_subscription_id, stripe_customer_id, plan_tier, status, current_period_start, current_period_end, cancel_at_period_end
			) VALUES (
				$1, 'sub_pro', 'cus_free', 'pro', 'active', $2, $3, false
			)
		`, workspaceA, time.Now().UTC(), time.Now().AddDate(0, 1, 0).UTC())
		require.NoError(t, err)

		// 1001st link should now succeed
		_, err = linkSvc.CreateLink(ctx, &workspaceA, payload)
		require.NoError(t, err)
	})

	t.Run("Cross-Tenant Isolation", func(t *testing.T) {
		workspaceB := uuid.New()
		_, err = dbPool.Exec(ctx, `INSERT INTO workspaces (id, name, stripe_customer_id) VALUES ($1, 'Workspace B', 'cus_b')`, workspaceB)
		require.NoError(t, err)

		// B is free, should be able to create link independently of A's 1001 links
		_, err = linkSvc.CreateLink(ctx, &workspaceB, payload)
		require.NoError(t, err)
	})
}
