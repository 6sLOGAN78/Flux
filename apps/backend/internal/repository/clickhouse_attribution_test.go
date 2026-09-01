package repository_test

import (
	"context"
	"testing"
	"time"

	"flux/apps/backend/internal/db"
	"flux/apps/backend/internal/model/analytics"
	"flux/apps/backend/internal/repository"
	pkgtesting "flux/apps/backend/internal/testing"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClickHouse_AttributionQuery_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping clickhouse integration test in short mode")
	}

	ctx := context.Background()
	chContainer, err := pkgtesting.SetupClickHouseContainer(ctx)
	require.NoError(t, err)
	defer chContainer.Terminate(context.Background())

	chConn, err := db.InitClickHouse(chContainer.Address)
	require.NoError(t, err)

	err = db.MigrateClickHouseSchema(ctx, chConn)
	require.NoError(t, err)

	workspaceID := uuid.New().String()
	workspaceB := uuid.New().String()

	// 1. Insert Analytics Events
	err = insertAnalyticsEvents(ctx, chConn, []analytics.AnalyticsEvent{
		{
			EventID:     "event-1",
			WorkspaceID: workspaceID,
			Timestamp:   time.Now().Add(-10 * time.Minute).UTC(),
			LinkID:      "link-1",
			CampaignID:  stringPtr("camp-1"),
			Referrer:    "google.com",
			UTMSource:   stringPtr("google"),
		},
		{
			EventID:     "event-2",
			WorkspaceID: workspaceID,
			Timestamp:   time.Now().Add(-5 * time.Minute).UTC(),
			LinkID:      "link-2",
			// No campaign ID for event-2
			Referrer: "twitter.com",
		},
		{
			EventID:     "event-3",
			WorkspaceID: workspaceB, // Different workspace
			Timestamp:   time.Now().Add(-2 * time.Minute).UTC(),
			LinkID:      "link-3",
		},
	})
	require.NoError(t, err)

	// 2. Insert Conversions
	err = insertConversions(ctx, chConn, []analytics.ConversionEvent{
		{
			ConversionID:   "conv-1",
			WorkspaceID:    workspaceID,
			Timestamp:      time.Now().UTC(),
			ConversionName: "signup",
			Revenue:        100.0,
			Currency:       "USD",
			ClickIDs:       []string{"event-1", "event-2", "event-3", "unknown-event"},
			VisitorID:      "vis-1",
		},
		{
			// Duplicate conversion to test Deduplication
			ConversionID:   "conv-1",
			WorkspaceID:    workspaceID,
			Timestamp:      time.Now().UTC(),
			ConversionName: "signup",
			Revenue:        100.0,
			Currency:       "USD",
			ClickIDs:       []string{"event-1", "event-2", "event-3", "unknown-event"},
			VisitorID:      "vis-1",
		},
	})
	require.NoError(t, err)

	// We need to wait a tiny bit for async merges/flushes if any, though ClickHouse is mostly synchronous on insert.
	time.Sleep(200 * time.Millisecond)

	repo := repository.NewClickHouseAnalyticsRepository(chConn)
	
	from := time.Now().Add(-1 * time.Hour)
	to := time.Now().Add(1 * time.Hour)
	
	conversions, err := repo.GetConversionsWithTouchpoints(ctx, workspaceID, from, to)
	require.NoError(t, err)

	// Should only return 1 deduplicated conversion
	require.Len(t, conversions, 1)
	
	conv := conversions[0]
	assert.Equal(t, 100.0, conv.Revenue)
	
	// Touchpoints should be exactly 2 (event-1 and event-2).
	// event-3 belongs to workspaceB so it's isolated.
	// unknown-event doesn't exist in analytics_events so it's ignored.
	require.Len(t, conv.Touchpoints, 2)
	
	// Check order (Chronological ascending)
	assert.Equal(t, repository.StringToUUID("event-1"), conv.Touchpoints[0].ID)
	assert.Equal(t, repository.StringToUUID("event-2"), conv.Touchpoints[1].ID)

	// Check mapping
	assert.Equal(t, repository.StringToUUID("camp-1"), conv.Touchpoints[0].CampaignID)
	assert.Equal(t, "camp-1", conv.Touchpoints[0].CampaignName)

	assert.Equal(t, repository.StringToUUID("link-2"), conv.Touchpoints[1].CampaignID) // Falls back to Link ID when missing
	assert.Equal(t, "link-2", conv.Touchpoints[1].CampaignName)

	// Try fetching for Workspace B
	conversionsB, err := repo.GetConversionsWithTouchpoints(ctx, workspaceB, from, to)
	require.NoError(t, err)
	assert.Len(t, conversionsB, 0)
}

func stringPtr(s string) *string {
	return &s
}

func insertAnalyticsEvents(ctx context.Context, conn driver.Conn, events []analytics.AnalyticsEvent) error {
	batch, err := conn.PrepareBatch(ctx, "INSERT INTO analytics_events")
	if err != nil {
		return err
	}
	for _, e := range events {
		err := batch.Append(
			e.EventID,
			string(analytics.EventTypeLinkRedirect),
			e.Timestamp,
			e.LinkID,
			e.WorkspaceID,
			e.ShortCode,
			e.Referrer,
			e.UserAgent,
			e.IPHash,
			e.CampaignID,
			e.UTMSource,
			e.UTMMedium,
			e.UTMCampaign,
			e.UTMTerm,
			e.UTMContent,
			e.CustomDomainID,
			nil, // hostname
		)
		if err != nil {
			return err
		}
	}
	return batch.Send()
}

func insertConversions(ctx context.Context, conn driver.Conn, events []analytics.ConversionEvent) error {
	batch, err := conn.PrepareBatch(ctx, "INSERT INTO conversions")
	if err != nil {
		return err
	}
	for _, e := range events {
		err := batch.Append(
			e.ConversionID,
			e.WorkspaceID,
			e.Timestamp,
			e.ConversionName,
			e.Revenue,
			e.Currency,
			e.ClickIDs,
			e.VisitorID,
		)
		if err != nil {
			return err
		}
	}
	return batch.Send()
}
