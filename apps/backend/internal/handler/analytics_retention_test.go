package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"flux/apps/backend/internal/database"
	"flux/apps/backend/internal/handler"
	"flux/apps/backend/internal/repository"
	pkgtesting "flux/apps/backend/internal/testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyticsRetention_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping analytics retention test in short mode")
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

	billingRepo := repository.NewBillingRepository(dbPool)
	mockProvider := &mockAnalyticsProvider{} // re-use from analytics_test.go

	analyticsH := handler.NewAnalyticsHandler(mockProvider, billingRepo)

	e := echo.New()

	// Workspace A (Free)
	workspaceA := uuid.New()
	_, err = dbPool.Exec(ctx, `INSERT INTO workspaces (id, name) VALUES ($1, 'Workspace Free')`, workspaceA)
	require.NoError(t, err)

	// Workspace B (Pro)
	workspaceB := uuid.New()
	_, err = dbPool.Exec(ctx, `INSERT INTO workspaces (id, name, stripe_customer_id) VALUES ($1, 'Workspace Pro', 'cus_pro')`, workspaceB)
	require.NoError(t, err)
	
	_, err = dbPool.Exec(ctx, `
		INSERT INTO subscriptions (
			workspace_id, stripe_subscription_id, stripe_customer_id, plan_tier, status, current_period_start, current_period_end, cancel_at_period_end
		) VALUES (
			$1, 'sub_pro', 'cus_pro', 'pro', 'active', $2, $3, false
		)
	`, workspaceB, time.Now().UTC(), time.Now().AddDate(0, 1, 0).UTC())
	require.NoError(t, err)

	// Test case: Request 365 days ago data
	fromParam := time.Now().UTC().AddDate(0, 0, -365).Format(time.RFC3339)

	t.Run("Free Workspace Retention Cap", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?from="+fromParam, nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("tenant_id", workspaceA)

		err := analyticsH.GetSummary(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		
		// The mockProvider.CapturedFrom should be forcefully snapped to 7 days ago!
		expectedBoundary := time.Now().UTC().AddDate(0, 0, -7)
		diff := mockProvider.CapturedFrom.Sub(expectedBoundary).Abs()
		assert.True(t, diff < 5*time.Second, "Free retention boundary not enforced: expected ~%v, got %v", expectedBoundary, mockProvider.CapturedFrom)
	})

	t.Run("Pro Workspace Retention Cap", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?from="+fromParam, nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("tenant_id", workspaceB)

		err := analyticsH.GetSummary(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		
		// The mockProvider.CapturedFrom should be forcefully snapped to 30 days ago for Pro!
		expectedBoundary := time.Now().UTC().AddDate(0, 0, -30)
		diff := mockProvider.CapturedFrom.Sub(expectedBoundary).Abs()
		assert.True(t, diff < 5*time.Second, "Pro retention boundary not enforced: expected ~%v, got %v", expectedBoundary, mockProvider.CapturedFrom)
	})
}
