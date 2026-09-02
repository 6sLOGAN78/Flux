package service_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"flux/apps/backend/internal/config"
	"flux/apps/backend/internal/database"
	"flux/apps/backend/internal/model/webhook"
	"flux/apps/backend/internal/repository"
	"flux/apps/backend/internal/service"
	pkgtesting "flux/apps/backend/internal/testing"
	"flux/apps/backend/internal/lib/utils"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateRetryDelay(t *testing.T) {
	initial := 5 * time.Second
	maxDelay := 1 * time.Hour

	// Attempt 1 -> ~5s
	delay1 := service.CalculateRetryDelay(1, initial, maxDelay)
	assert.GreaterOrEqual(t, float64(delay1), float64(4*time.Second))
	assert.LessOrEqual(t, float64(delay1), float64(6*time.Second))

	// Attempt 2 -> ~10s
	delay2 := service.CalculateRetryDelay(2, initial, maxDelay)
	assert.GreaterOrEqual(t, float64(delay2), float64(8*time.Second))
	assert.LessOrEqual(t, float64(delay2), float64(12*time.Second))

	// Attempt 3 -> ~20s
	delay3 := service.CalculateRetryDelay(3, initial, maxDelay)
	assert.GreaterOrEqual(t, float64(delay3), float64(16*time.Second))
	assert.LessOrEqual(t, float64(delay3), float64(24*time.Second))

	// Attempt 15 -> should cap at maxDelay (1 hour)
	delay15 := service.CalculateRetryDelay(15, initial, maxDelay)
	assert.GreaterOrEqual(t, float64(delay15), float64(48*time.Minute))
	assert.LessOrEqual(t, float64(delay15), float64(72*time.Minute))
}

func TestIsRetryableError(t *testing.T) {
	// Status code checks
	assert.False(t, service.IsRetryableError(400, nil))
	assert.False(t, service.IsRetryableError(404, nil))
	assert.False(t, service.IsRetryableError(410, nil))
	assert.True(t, service.IsRetryableError(429, nil))
	assert.True(t, service.IsRetryableError(500, nil))
	assert.True(t, service.IsRetryableError(502, nil))
	assert.False(t, service.IsRetryableError(501, nil))

	// Explicit network error
	assert.True(t, service.IsRetryableError(0, errors.New("connection reset by peer")))
	
	// SSRF is NOT retryable
	assert.False(t, service.IsRetryableError(0, utils.ErrSSRFBlocked))
}

func TestWebhookRetryWorker_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping webhook retry worker integration test in short mode")
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

	repo := repository.NewWebhookRepository(dbPool)

	workspaceID := uuid.New()
	_, err = dbPool.Exec(ctx, `INSERT INTO workspaces (id, name) VALUES ($1, 'Workspace Retry')`, workspaceID)
	require.NoError(t, err)

	var requestCount int

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		// Fail the first request, succeed the second
		if requestCount == 1 {
			w.WriteHeader(http.StatusServiceUnavailable) // 503 -> Retryable
			return
		}
		if requestCount == 2 {
			w.WriteHeader(http.StatusOK) // 200 -> Success
			return
		}
	}))
	defer ts.Close()

	wh, err := repo.CreateWebhook(ctx, &webhook.Webhook{
		WorkspaceID: workspaceID,
		EndpointURL: ts.URL,
		Secret:      "whsec_test_secret",
		Active:      true,
		Events:      []string{"test.event"},
	})
	require.NoError(t, err)

	// Insert a delivery that needs retrying right away!
	d1 := &repository.WebhookDelivery{
		ID:             uuid.New(),
		WebhookID:      wh.ID,
		EventID:        "evt_retry_1",
		Status:         "retrying",
		AttemptCount:   1, // It failed once initially
		Payload:        []byte(`{"msg":"retry_me"}`),
	}
	// We want it to be picked up immediately
	past := time.Now().Add(-5 * time.Minute)
	d1.NextAttemptAt = &past
	err = repo.RecordDelivery(ctx, d1)
	require.NoError(t, err)

	cfg := &config.WebhookConfig{
		WorkerConcurrency: 2,
		DeliveryTimeout:   "2s",
		MaxRetries:        3,
		RetryInitialDelay: "1s",
		RetryMaxDelay:     "5s",
		RetryPollInterval: "1s",
		RetryConcurrency:  2,
	}

	worker := service.NewWebhookRetryWorker(repo, cfg)
	worker.SetHTTPClient(ts.Client()) // Bypass SSRF for test
	worker.Start()

	// Wait for processing logic
	var status string
	var finalAttempts int
	assert.Eventually(t, func() bool {
		err := dbPool.QueryRow(ctx, "SELECT status, attempt_count FROM webhook_deliveries WHERE id = $1", d1.ID).Scan(&status, &finalAttempts)
		return err == nil && status == "success" && finalAttempts == 2
	}, 10*time.Second, 500*time.Millisecond)

	worker.Stop(1 * time.Second)

	// Now let's test Dead-Letter
	d2 := &repository.WebhookDelivery{
		ID:             uuid.New(),
		WebhookID:      wh.ID,
		EventID:        "evt_retry_dead",
		Status:         "retrying",
		AttemptCount:   2, // It already failed twice, max is 3. This attempt will make it 3 -> dead letter.
		Payload:        []byte(`{"msg":"die"}`),
	}
	d2.NextAttemptAt = &past
	err = repo.RecordDelivery(ctx, d2)
	require.NoError(t, err)

	// Reset mock server to always fail 500
	ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	worker2 := service.NewWebhookRetryWorker(repo, cfg)
	worker2.SetHTTPClient(ts.Client())
	worker2.Start()
	
	assert.Eventually(t, func() bool {
		err := dbPool.QueryRow(ctx, "SELECT status, attempt_count FROM webhook_deliveries WHERE id = $1", d2.ID).Scan(&status, &finalAttempts)
		return err == nil && status == "dead_letter" && finalAttempts == 3
	}, 10*time.Second, 500*time.Millisecond)

	worker2.Stop(1 * time.Second)
}
