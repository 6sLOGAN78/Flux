package analytics_test

import (
	"context"
	"testing"
	"time"

	"flux/apps/backend/internal/modules/analytics"

	"github.com/google/uuid"
)

func TestClickHouseBatchWriter_BatchingAndFlush(t *testing.T) {
	mockStore := analytics.NewMockClickHouseStore()
	writer := analytics.NewClickHouseBatchWriter(mockStore, 10, 50*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	writer.Start(ctx)

	linkID := uuid.New()
	userID := uuid.New()

	for i := 0; i < 15; i++ {
		writer.Add(analytics.ClickEvent{
			ID:             uuid.New(),
			LinkID:         linkID,
			UserID:         userID,
			Timestamp:      time.Now(),
			IPAddress:      "192.168.1.1",
			CountryCode:    "US",
			Browser:        "Chrome",
			OS:             "macOS",
			DeviceType:     "desktop",
			ResponseTimeMS: 45,
		})
	}

	// Wait for automatic batch flush or call Flush
	time.Sleep(100 * time.Millisecond)
	_ = writer.Flush(context.Background())

	events := mockStore.GetEvents()
	if len(events) != 15 {
		t.Fatalf("expected 15 events in ClickHouse store, got %d", len(events))
	}
}

func TestClickHouseBatchWriter_QueryTimeSeries(t *testing.T) {
	mockStore := analytics.NewMockClickHouseStore()
	linkID := uuid.New()
	userID := uuid.New()

	now := time.Now()
	mockStore.AddEvents([]analytics.ClickEvent{
		{
			ID:        uuid.New(),
			LinkID:    linkID,
			UserID:    userID,
			Timestamp: now.Add(-30 * time.Minute),
		},
		{
			ID:        uuid.New(),
			LinkID:    linkID,
			UserID:    userID,
			Timestamp: now.Add(-10 * time.Minute),
		},
	})

	points, err := mockStore.QueryTimeSeries(context.Background(), linkID.String(), now.Add(-1*time.Hour), now, "hour")
	if err != nil {
		t.Fatalf("expected query to succeed, got error: %v", err)
	}

	if len(points) == 0 {
		t.Fatalf("expected time series data points, got empty slice")
	}
}
