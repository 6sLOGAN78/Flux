package service_test

import (
	"context"
	"testing"
	"time"

	"flux/apps/backend/internal/db"
	"flux/apps/backend/internal/model/analytics"
	"flux/apps/backend/internal/service"
	pkgtesting "flux/apps/backend/internal/testing"

	"github.com/redis/go-redis/v9"
)

func TestClickHouseConsumer_E2E_Success(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// 1. Setup Redis
	redisContainer, err := pkgtesting.SetupRedisContainer(ctx)
	if err != nil {
		t.Fatalf("failed to start redis container: %v", err)
	}
	defer redisContainer.Terminate(context.Background())

	redisClient := redis.NewClient(&redis.Options{Addr: redisContainer.Address})
	defer redisClient.Close()

	// 2. Setup ClickHouse
	chContainer, err := pkgtesting.SetupClickHouseContainer(ctx)
	if err != nil {
		t.Fatalf("failed to start clickhouse container: %v", err)
	}
	defer chContainer.Terminate(context.Background())

	chConn, err := db.InitClickHouse(chContainer.Address)
	if err != nil {
		t.Fatalf("failed to init clickhouse: %v", err)
	}
	
	err = db.MigrateClickHouseSchema(ctx, chConn)
	if err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}

	// 3. Start Consumer
	streamName := "analytics:test:ch"
	consumer := service.NewRedisAnalyticsConsumer(redisClient, chConn, streamName)
	consumer.Start()

	// 4. Start Publisher
	publisher := service.NewRedisAnalyticsPublisher(redisClient, streamName, 100)
	publisher.Start()

	// 5. Emit Events (2 for distinct workspaces)
	event1 := &analytics.AnalyticsEvent{
		EventID:     "ch_evt_1",
		EventType:   analytics.EventTypeLinkRedirect,
		Timestamp:   time.Now().UTC().Truncate(time.Millisecond),
		LinkID:      "link_1",
		WorkspaceID: "ws_A",
	}

	event2 := &analytics.AnalyticsEvent{
		EventID:     "ch_evt_2",
		EventType:   analytics.EventTypeLinkRedirect,
		Timestamp:   time.Now().UTC().Truncate(time.Millisecond),
		LinkID:      "link_2",
		WorkspaceID: "ws_B",
	}

	_ = publisher.PublishEvent(ctx, event1)
	_ = publisher.PublishEvent(ctx, event2)

	// Give it time to flush and batch
	time.Sleep(3 * time.Second)

	publisher.Stop(2 * time.Second)
	consumer.Stop(2 * time.Second)

	// 6. Query ClickHouse Directly
	rows, err := chConn.Query(ctx, "SELECT event_id, workspace_id FROM analytics_events ORDER BY event_id ASC")
	if err != nil {
		t.Fatalf("failed to query clickhouse: %v", err)
	}
	defer rows.Close()

	type resultRow struct {
		EventID     string
		WorkspaceID string
	}
	var results []resultRow

	for rows.Next() {
		var r resultRow
		if err := rows.Scan(&r.EventID, &r.WorkspaceID); err != nil {
			t.Fatalf("failed to scan row: %v", err)
		}
		results = append(results, r)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 events in clickhouse, got %d", len(results))
	}

	if results[0].EventID != "ch_evt_1" || results[0].WorkspaceID != "ws_A" {
		t.Errorf("expected ch_evt_1/ws_A, got %v", results[0])
	}
	if results[1].EventID != "ch_evt_2" || results[1].WorkspaceID != "ws_B" {
		t.Errorf("expected ch_evt_2/ws_B, got %v", results[1])
	}
	
	// 7. Verify ACK
	// Pending list should be empty
	pending, err := redisClient.XPending(ctx, streamName, "analytics-clickhouse").Result()
	if err != nil {
		t.Fatalf("failed to check pending list: %v", err)
	}
	if pending.Count != 0 {
		t.Fatalf("expected 0 pending messages, got %d. XACK failed.", pending.Count)
	}
}
