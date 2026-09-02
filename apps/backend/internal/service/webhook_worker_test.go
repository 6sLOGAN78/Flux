package service_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"flux/apps/backend/internal/database"
	"flux/apps/backend/internal/model/analytics"
	"flux/apps/backend/internal/model/webhook"
	"flux/apps/backend/internal/repository"
	"flux/apps/backend/internal/service"
	pkgtesting "flux/apps/backend/internal/testing"
	"flux/apps/backend/internal/lib/utils"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookWorker_Signing(t *testing.T) {
	secret := "whsec_1234567890abcdef"
	payload := []byte(`{"id":"evt_1","type":"link.redirect"}`)

	sig1 := utils.GenerateHMACSHA256(secret, payload)
	sig2 := utils.GenerateHMACSHA256(secret, payload)
	
	// Same
	assert.Equal(t, sig1, sig2)
	assert.True(t, utils.VerifyHMACSHA256(secret, payload, sig1))

	// Diff payload
	payload2 := []byte(`{"id":"evt_2","type":"link.redirect"}`)
	sig3 := utils.GenerateHMACSHA256(secret, payload2)
	assert.NotEqual(t, sig1, sig3)
	assert.False(t, utils.VerifyHMACSHA256(secret, payload2, sig1))

	// Diff secret
	secret2 := "whsec_fedcba0987654321"
	sig4 := utils.GenerateHMACSHA256(secret2, payload)
	assert.NotEqual(t, sig1, sig4)
	assert.False(t, utils.VerifyHMACSHA256(secret2, payload, sig1))
}

func TestWebhookWorker_SSRF(t *testing.T) {
	client := utils.SafeHTTPClient(2 * time.Second)

	// Test private IP
	_, err := client.Get("http://127.0.0.1:8080")
	require.Error(t, err)
	assert.ErrorIs(t, err, utils.ErrSSRFBlocked)

	// Test localhost
	_, err = client.Get("http://localhost:8080")
	require.Error(t, err)
	assert.ErrorIs(t, err, utils.ErrSSRFBlocked)

	// Test link-local
	_, err = client.Get("http://169.254.169.254")
	require.Error(t, err)
	assert.ErrorIs(t, err, utils.ErrSSRFBlocked)
}

func TestWebhookWorker_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping webhook worker integration test in short mode")
	}

	ctx := context.Background()
	pgContainer, err := pkgtesting.SetupPostgresContainer(ctx)
	require.NoError(t, err)
	defer pgContainer.Terminate(ctx)

	redisContainer, err := pkgtesting.SetupRedisContainer(ctx)
	require.NoError(t, err)
	defer redisContainer.Terminate(ctx)

	dbPool, err := database.InitDBPool(ctx, pgContainer.DSN)
	require.NoError(t, err)
	defer dbPool.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: redisContainer.Address})
	defer redisClient.Close()

	logger := zerolog.Nop()
	err = database.MigrateDSN(ctx, &logger, pgContainer.DSN)
	require.NoError(t, err)

	repo := repository.NewWebhookRepository(dbPool)

	workspaceA := uuid.New()
	_, err = dbPool.Exec(ctx, `INSERT INTO workspaces (id, name) VALUES ($1, 'Workspace A')`, workspaceA)
	require.NoError(t, err)

	workspaceB := uuid.New()
	_, err = dbPool.Exec(ctx, `INSERT INTO workspaces (id, name) VALUES ($1, 'Workspace B')`, workspaceB)
	require.NoError(t, err)

	var wg sync.WaitGroup
	var receivedPayloads []webhook.WebhookEventPayload
	var mu sync.Mutex

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var p webhook.WebhookEventPayload
		json.NewDecoder(r.Body).Decode(&p)
		mu.Lock()
		receivedPayloads = append(receivedPayloads, p)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
		wg.Done()
	}))
	defer ts.Close()

	// Webhook for Workspace A (subbed to link.redirect)
	secA, _ := utils.GenerateWebhookSecret()
	whA := &webhook.Webhook{
		WorkspaceID: workspaceA,
		EndpointURL: ts.URL,
		Secret:      secA,
		Active:      true,
		Events:      []string{"link.redirect"},
	}
	_, err = repo.CreateWebhook(ctx, whA)
	require.NoError(t, err)

	// Webhook for Workspace B (subbed to conversion)
	secB, _ := utils.GenerateWebhookSecret()
	whB := &webhook.Webhook{
		WorkspaceID: workspaceB,
		EndpointURL: ts.URL,
		Secret:      secB,
		Active:      true,
		Events:      []string{"conversion"},
	}
	_, err = repo.CreateWebhook(ctx, whB)
	require.NoError(t, err)

	worker := service.NewWebhookWorker(redisClient, repo, "analytics:events", 5, 2*time.Second)
	worker.SetHTTPClient(ts.Client()) // bypass SSRF check for test server
	worker.Start()
	defer worker.Stop(1 * time.Second)

	// Publish Event for Workspace A (link.redirect)
	evt1 := analytics.AnalyticsEvent{
		EventID:     "evt_1",
		EventType:   "link.redirect",
		Timestamp:   time.Now(),
		WorkspaceID: workspaceA.String(),
	}
	data, _ := json.Marshal(evt1)
	redisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: "analytics:events",
		Values: map[string]interface{}{"payload": data},
	})

	// Publish Event for Workspace B (link.redirect - but B is only subbed to conversion!)
	evt2 := analytics.AnalyticsEvent{
		EventID:     "evt_2",
		EventType:   "link.redirect",
		Timestamp:   time.Now(),
		WorkspaceID: workspaceB.String(),
	}
	data2, _ := json.Marshal(evt2)
	redisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: "analytics:events",
		Values: map[string]interface{}{"payload": data2},
	})

	// Publish Event for Workspace B (conversion)
	evt3 := analytics.AnalyticsEvent{
		EventID:     "evt_3",
		EventType:   "conversion",
		Timestamp:   time.Now(),
		WorkspaceID: workspaceB.String(),
	}
	data3, _ := json.Marshal(evt3)
	redisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: "analytics:events",
		Values: map[string]interface{}{"payload": data3},
	})

	wg.Add(2) // We expect exactly 2 deliveries: evt1 and evt3. evt2 should be skipped due to event mismatch.

	// Wait for deliveries
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for webhooks to be delivered")
	}

	mu.Lock()
	defer mu.Unlock()
	
	time.Sleep(100 * time.Millisecond)
	assert.Len(t, receivedPayloads, 2)
	
	ids := map[string]bool{}
	for _, p := range receivedPayloads {
		ids[p.ID] = true
	}
	assert.True(t, ids["evt_1"])
	assert.True(t, ids["evt_3"])
	assert.False(t, ids["evt_2"]) // Successfully filtered!
	
	// Check that deliveries were recorded in DB
	var count int
	err = dbPool.QueryRow(ctx, "SELECT COUNT(*) FROM webhook_deliveries").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 2, count, "should have 2 delivery records in database")
}
