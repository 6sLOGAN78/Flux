package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"flux/apps/backend/internal/config"
	"flux/apps/backend/internal/database"
	"flux/apps/backend/internal/handler"
	"flux/apps/backend/internal/repository"
	pkgtesting "flux/apps/backend/internal/testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v78/webhook"
)

func TestStripeWebhookHandler(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping webhook integration test in short mode")
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

	repo := repository.NewBillingRepository(dbPool)
	cfg := &config.Config{Stripe: config.StripeConfig{WebhookSecret: "whsec_test"}}
	whHandler := handler.NewStripeWebhookHandler(dbPool, repo, cfg)

	e := echo.New()

	// 1. Setup workspace and user
	workspaceID := uuid.New()
	_, err = dbPool.Exec(ctx, `INSERT INTO workspaces (id, name, stripe_customer_id) VALUES ($1, 'Test', 'cus_123')`, workspaceID)
	require.NoError(t, err)

	// Helper to generate signed payload
	generatePayload := func(eventType, eventID, subID, custID string) ([]byte, string) {
		payloadMap := map[string]interface{}{
			"id": eventID,
			"type": eventType,
			"api_version": "2024-04-10",
			"data": map[string]interface{}{
				"object": map[string]interface{}{
					"id": subID,
					"customer": custID,
					"status": "active",
					"current_period_start": time.Now().Unix(),
					"current_period_end": time.Now().AddDate(0, 1, 0).Unix(),
				},
			},
		}
		payloadJSON, _ := json.Marshal(payloadMap)

		sig := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
			Payload: payloadJSON,
			Secret:  "whsec_test",
			Timestamp: time.Now(),
		})

		return payloadJSON, sig.Header
	}

	t.Run("Missing Signature", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/stripe", bytes.NewReader([]byte("{}")))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		
		err := whHandler.HandleWebhook(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Missing Stripe-Signature")
	})

	t.Run("Invalid Signature", func(t *testing.T) {
		payloadJSON, _ := generatePayload("customer.subscription.created", "evt_1", "sub_1", "cus_123")
		
		req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/stripe", bytes.NewReader(payloadJSON))
		req.Header.Set("Stripe-Signature", "t=123,v1=invalid")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		
		err := whHandler.HandleWebhook(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Invalid signature")
	})

	t.Run("Valid customer.subscription.created", func(t *testing.T) {
		payloadJSON, sig := generatePayload("customer.subscription.created", "evt_1", "sub_1", "cus_123")
		
		req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/stripe", bytes.NewReader(payloadJSON))
		req.Header.Set("Stripe-Signature", sig)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		
		err := whHandler.HandleWebhook(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		// Verify DB state
		var count int
		dbPool.QueryRow(ctx, "SELECT count(*) FROM subscriptions WHERE stripe_subscription_id = 'sub_1'").Scan(&count)
		assert.Equal(t, 1, count)
	})

	t.Run("Idempotency: Duplicate Event", func(t *testing.T) {
		payloadJSON, sig := generatePayload("customer.subscription.created", "evt_1", "sub_1", "cus_123") // Reusing evt_1
		
		req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/stripe", bytes.NewReader(payloadJSON))
		req.Header.Set("Stripe-Signature", sig)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		
		err := whHandler.HandleWebhook(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Already processed")
	})

	t.Run("Unknown Customer", func(t *testing.T) {
		payloadJSON, sig := generatePayload("customer.subscription.created", "evt_3", "sub_3", "cus_999") // cus_999 doesn't exist
		
		req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/stripe", bytes.NewReader(payloadJSON))
		req.Header.Set("Stripe-Signature", sig)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		
		err := whHandler.HandleWebhook(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code) // Expected to return 200 so Stripe doesn't infinitely retry
		assert.Contains(t, rec.Body.String(), "Customer unknown")
	})

	t.Run("Tenant Mismatch (Security Check)", func(t *testing.T) {
		// First assign sub_2 to cus_123 / Workspace 1
		payloadJSON, sig := generatePayload("customer.subscription.created", "evt_4", "sub_2", "cus_123")
		
		req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/stripe", bytes.NewReader(payloadJSON))
		req.Header.Set("Stripe-Signature", sig)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		whHandler.HandleWebhook(c)

		// Create another workspace with cus_456
		workspaceID2 := uuid.New()
		_, err = dbPool.Exec(ctx, `INSERT INTO workspaces (id, name, stripe_customer_id) VALUES ($1, 'Test 2', 'cus_456')`, workspaceID2)
		require.NoError(t, err)

		// Attempt to update sub_2 but via cus_456
		payloadJSON2, sig2 := generatePayload("customer.subscription.updated", "evt_5", "sub_2", "cus_456")
		
		req2 := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/stripe", bytes.NewReader(payloadJSON2))
		req2.Header.Set("Stripe-Signature", sig2)
		rec2 := httptest.NewRecorder()
		c2 := e.NewContext(req2, rec2)
		
		err = whHandler.HandleWebhook(c2)
		require.NoError(t, err)
		assert.Equal(t, http.StatusConflict, rec2.Code)
		assert.Contains(t, rec2.Body.String(), "Cross-tenant mismatch")
	})
}
func TestStripeWebhookHandler_Concurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping webhook concurrency test in short mode")
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

	repo := repository.NewBillingRepository(dbPool)
	cfg := &config.Config{Stripe: config.StripeConfig{WebhookSecret: "whsec_test"}}
	whHandler := handler.NewStripeWebhookHandler(dbPool, repo, cfg)

	e := echo.New()

	// 1. Setup workspace and user
	workspaceID := uuid.New()
	_, err = dbPool.Exec(ctx, `INSERT INTO workspaces (id, name, stripe_customer_id) VALUES ($1, 'Test Concurrent', 'cus_concurrent')`, workspaceID)
	require.NoError(t, err)

	generatePayload := func(eventType, eventID, subID, custID string) ([]byte, string) {
		payloadMap := map[string]interface{}{
			"id": eventID,
			"type": eventType,
			"api_version": "2024-04-10",
			"data": map[string]interface{}{
				"object": map[string]interface{}{
					"id": subID,
					"customer": custID,
					"status": "active",
					"current_period_start": time.Now().Unix(),
					"current_period_end": time.Now().AddDate(0, 1, 0).Unix(),
				},
			},
		}
		payloadJSON, _ := json.Marshal(payloadMap)

		sig := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
			Payload: payloadJSON,
			Secret:  "whsec_test",
			Timestamp: time.Now(),
		})

		return payloadJSON, sig.Header
	}

	payloadJSON, sig := generatePayload("customer.subscription.created", "evt_concurrent", "sub_concurrent", "cus_concurrent")

	// Fire 10 concurrent webhook requests
	const numRequests = 10
	results := make(chan int, numRequests)
	
	for i := 0; i < numRequests; i++ {
		go func() {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/stripe", bytes.NewReader(payloadJSON))
			req.Header.Set("Stripe-Signature", sig)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			
			err := whHandler.HandleWebhook(c)
			require.NoError(t, err)
			results <- rec.Code
		}()
	}

	
	for i := 0; i < numRequests; i++ {
		code := <-results
		// Both 200 OK because already processed also returns 200 to Stripe to ack
		if code == http.StatusOK {
			// We can't easily distinguish from response code alone since both return 200
			// but we can query DB
		}
	}

	var count int
	err = dbPool.QueryRow(ctx, "SELECT count(*) FROM subscriptions WHERE stripe_subscription_id = 'sub_concurrent'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Subscription should be inserted exactly once")

	err = dbPool.QueryRow(ctx, "SELECT count(*) FROM stripe_events WHERE id = 'evt_concurrent'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Stripe event should be recorded exactly once")
}
