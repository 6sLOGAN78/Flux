package service_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"flux/apps/backend/internal/model/analytics"
	"flux/apps/backend/internal/service"
	pkgtesting "flux/apps/backend/internal/testing"

	"github.com/redis/go-redis/v9"
)

func TestRedisAnalyticsPublisher_Success(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	redisContainer, err := pkgtesting.SetupRedisContainer(ctx)
	if err != nil {
		t.Fatalf("failed to start redis container: %v", err)
	}
	defer redisContainer.Terminate(context.Background())

	client := redis.NewClient(&redis.Options{
		Addr: redisContainer.Address,
	})
	defer client.Close()

	streamName := "analytics:events:test"
	pub := service.NewRedisAnalyticsPublisher(client, streamName, 100)
	pub.Start()
	
	// Create canonical event
	event := &analytics.AnalyticsEvent{
		EventID:     "evt_123",
		EventType:   analytics.EventTypeLinkRedirect,
		Timestamp:   time.Now().UTC(),
		LinkID:      "link_456",
		WorkspaceID: "ws_789",
		ShortCode:   "AcMe12",
		Referrer:    "https://example.com",
		UserAgent:   "TestAgent/1.0",
		IPHash:      "abcd1234hash",
	}

	err = pub.PublishEvent(ctx, event)
	if err != nil {
		t.Fatalf("unexpected error publishing event: %v", err)
	}

	// Give the async worker a moment to process the queue
	time.Sleep(100 * time.Millisecond)
	
	pub.Stop(2 * time.Second)

	// Verify event exists in Redis Stream
	messages, err := client.XRange(ctx, streamName, "-", "+").Result()
	if err != nil {
		t.Fatalf("failed to read from redis stream: %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("expected 1 message in stream, got %d", len(messages))
	}

	payloadStr, ok := messages[0].Values["payload"].(string)
	if !ok {
		t.Fatalf("expected payload to be string")
	}

	var storedEvent analytics.AnalyticsEvent
	if err := json.Unmarshal([]byte(payloadStr), &storedEvent); err != nil {
		t.Fatalf("failed to unmarshal stored event: %v", err)
	}

	// Verify Serialization
	if storedEvent.EventID != event.EventID {
		t.Errorf("expected EventID %s, got %s", event.EventID, storedEvent.EventID)
	}
	if storedEvent.WorkspaceID != event.WorkspaceID {
		t.Errorf("expected WorkspaceID %s, got %s", event.WorkspaceID, storedEvent.WorkspaceID)
	}
	if storedEvent.IPHash != event.IPHash {
		t.Errorf("expected IPHash %s, got %s", event.IPHash, storedEvent.IPHash)
	}
}

func TestRedisAnalyticsPublisher_BufferFullFailureIsolation(t *testing.T) {
	// If Redis is completely hung/unavailable, the queue fills up.
	// We want to ensure PublishEvent does not block indefinitely and drops safely.
	
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:9999", // Invalid port
	})

	// Create publisher with very small buffer (1)
	pub := service.NewRedisAnalyticsPublisher(client, "test:stream", 1)
	// We DO NOT call pub.Start() to intentionally leave the worker unstarted, which simulates a completely stuck/blocked worker.

	ctx := context.Background()

	// Fill the buffer
	err1 := pub.PublishEvent(ctx, &analytics.AnalyticsEvent{EventID: "1"})
	if err1 != nil {
		t.Fatalf("first publish should succeed filling the buffer")
	}

	// Next publish should drop instantly without error and without blocking
	done := make(chan struct{})
	go func() {
		err2 := pub.PublishEvent(ctx, &analytics.AnalyticsEvent{EventID: "2"})
		if err2 != nil {
			t.Errorf("expected nil error (drop), got %v", err2)
		}
		close(done)
	}()

	select {
	case <-done:
		// Success, didn't block
	case <-time.After(1 * time.Second):
		t.Fatalf("PublishEvent blocked on full buffer instead of dropping")
	}
}
