package analytics_test

import (
	"context"
	"testing"
	"time"

	"flux/apps/backend/internal/modules/analytics"
)

// MockQueueProducer tracks published click events in memory for unit testing
type MockQueueProducer struct {
	events []*analytics.ClickEvent
}

func (m *MockQueueProducer) Publish(ctx context.Context, event *analytics.ClickEvent) error {
	m.events = append(m.events, event)
	return nil
}

func TestClickCollector_CollectAsync(t *testing.T) {
	producer := &MockQueueProducer{events: make([]*analytics.ClickEvent, 0)}
	collector := analytics.NewAsyncCollector(producer, 100)
	collector.Start(context.Background())
	defer collector.Stop()

	event := &analytics.ClickEvent{
		ID:        "evt_001",
		LinkID:    "link_123",
		Slug:      "openai",
		Timestamp: time.Now(),
		IPAddress: "192.168.1.1",
		UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
		Referrer:  "https://google.com",
	}

	// Fire async collection
	collector.CollectAsync(event)

	// Wait briefly for worker processing
	time.Sleep(50 * time.Millisecond)

	if len(producer.events) != 1 {
		t.Fatalf("expected 1 event published, got %d", len(producer.events))
	}

	pubEvent := producer.events[0]
	if pubEvent.Slug != "openai" {
		t.Errorf("expected slug 'openai', got '%s'", pubEvent.Slug)
	}

	if pubEvent.IPAddress != "192.168.1.1" {
		t.Errorf("expected IP '192.168.1.1', got '%s'", pubEvent.IPAddress)
	}
}

func TestClickCollector_NonBlocking(t *testing.T) {
	producer := &MockQueueProducer{events: make([]*analytics.ClickEvent, 0)}
	collector := analytics.NewAsyncCollector(producer, 10)
	collector.Start(context.Background())
	defer collector.Stop()

	start := time.Now()
	for i := 0; i < 50; i++ {
		collector.CollectAsync(&analytics.ClickEvent{
			ID:     "evt_bench",
			Slug:   "bench",
			LinkID: "link_bench",
		})
	}
	duration := time.Since(start)

	// CollectAsync must return virtually instantaneously (< 5ms)
	if duration > 10*time.Millisecond {
		t.Errorf("CollectAsync blocked main execution thread! Taken: %v", duration)
	}
}
