package repository_test

import (
	"context"
	"testing"
	"time"

	"flux/apps/backend/internal/db"
	"flux/apps/backend/internal/model/analytics"
	"flux/apps/backend/internal/repository"
	pkgtesting "flux/apps/backend/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClickHouseAnalyticsRepository_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	chContainer, err := pkgtesting.SetupClickHouseContainer(ctx)
	if err != nil {
		t.Fatalf("failed to setup clickhouse container: %v", err)
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

	repo := repository.NewClickHouseAnalyticsRepository(chConn)

	now := time.Now().UTC()
	wsA := "workspace_A"
	wsB := "workspace_B"
	
	// Insert test data
	batch, err := chConn.PrepareBatch(ctx, "INSERT INTO analytics_events")
	if err != nil {
		t.Fatalf("failed to prepare batch: %v", err)
	}
	
	events := []analytics.AnalyticsEvent{
		// Workspace A: 2 distinct events, 1 duplicate event (same event_id), 1 different ip
		{EventID: "evt_1", EventType: analytics.EventTypeLinkRedirect, Timestamp: now, LinkID: "link_1", WorkspaceID: wsA, ShortCode: "a1", Referrer: "google.com", IPHash: "ip_1"},
		{EventID: "evt_1", EventType: analytics.EventTypeLinkRedirect, Timestamp: now, LinkID: "link_1", WorkspaceID: wsA, ShortCode: "a1", Referrer: "google.com", IPHash: "ip_1"}, // Duplicate!
		{EventID: "evt_2", EventType: analytics.EventTypeLinkRedirect, Timestamp: now, LinkID: "link_2", WorkspaceID: wsA, ShortCode: "a2", Referrer: "twitter.com", IPHash: "ip_2"},
		// Old event out of time range
		{EventID: "evt_old", EventType: analytics.EventTypeLinkRedirect, Timestamp: now.Add(-40 * 24 * time.Hour), LinkID: "link_1", WorkspaceID: wsA, ShortCode: "a1", Referrer: "google.com", IPHash: "ip_1"},
		
		// Workspace B: 1 event
		{EventID: "evt_3", EventType: analytics.EventTypeLinkRedirect, Timestamp: now, LinkID: "link_3", WorkspaceID: wsB, ShortCode: "b1", Referrer: "google.com", IPHash: "ip_3"},
	}
	
	for _, e := range events {
		err := batch.Append(
			e.EventID, 
			string(e.EventType), 
			e.Timestamp, 
			e.LinkID, 
			e.WorkspaceID, 
			e.ShortCode, 
			e.Referrer, 
			e.UserAgent, 
			e.IPHash,
			(*string)(nil), // CampaignID
			(*string)(nil), // UTMSource
			(*string)(nil), // UTMMedium
			(*string)(nil), // UTMCampaign
			(*string)(nil), // UTMTerm
			(*string)(nil), // UTMContent
		)
		if err != nil {
			t.Fatalf("failed to append batch: %v", err)
		}
	}
	if err := batch.Send(); err != nil {
		t.Fatalf("failed to send batch: %v", err)
	}

	// Test 1: Workspace Isolation & Deduplication (Workspace A)
	from := now.Add(-24 * time.Hour)
	to := now.Add(24 * time.Hour)
	
	summaryA, err := repo.GetSummary(ctx, wsA, from, to)
	if err != nil {
		t.Fatalf("GetSummary failed: %v", err)
	}
	
	if summaryA.TotalClicks != 2 { // evt_1 and evt_2
		t.Errorf("expected 2 total clicks (deduplicated), got %d", summaryA.TotalClicks)
	}
	if summaryA.UniqueVisitors != 2 { // ip_1 and ip_2
		t.Errorf("expected 2 unique visitors, got %d", summaryA.UniqueVisitors)
	}
	
	// Test 2: Workspace Isolation (Workspace B)
	summaryB, err := repo.GetSummary(ctx, wsB, from, to)
	if err != nil {
		t.Fatalf("GetSummary failed: %v", err)
	}
	if summaryB.TotalClicks != 1 {
		t.Errorf("expected 1 total click for B, got %d", summaryB.TotalClicks)
	}
	
	// Test 3: Date Filtering
	fromOld := now.Add(-50 * 24 * time.Hour)
	toOld := now.Add(-30 * 24 * time.Hour)
	summaryOld, err := repo.GetSummary(ctx, wsA, fromOld, toOld)
	if err != nil {
		t.Fatalf("GetSummary old failed: %v", err)
	}
	if summaryOld.TotalClicks != 1 { // evt_old
		t.Errorf("expected 1 click in old range, got %d", summaryOld.TotalClicks)
	}
	
	// Test 4: Top Links
	topLinks, err := repo.GetTopLinks(ctx, wsA, from, to, 10)
	if err != nil {
		t.Fatalf("GetTopLinks failed: %v", err)
	}
	if len(topLinks.Data) != 2 {
		t.Fatalf("expected 2 top links, got %d", len(topLinks.Data))
	}
	// Deduplication should ensure link_1 has 1 click, not 2
	for _, l := range topLinks.Data {
		if l.LinkID == "link_1" && l.Clicks != 1 {
			t.Errorf("expected link_1 to have 1 click (deduplicated), got %d", l.Clicks)
		}
	}
	
	// Test 5: Referrers
	referrers, err := repo.GetReferrers(ctx, wsA, from, to, 10)
	if err != nil {
		t.Fatalf("GetReferrers failed: %v", err)
	}
	if len(referrers.Data) != 2 {
		t.Fatalf("expected 2 referrers, got %d", len(referrers.Data))
	}
}

func TestClickHouseAnalyticsRepository_CampaignUTMAttribution(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	chContainer, err := pkgtesting.SetupClickHouseContainer(ctx)
	if err != nil {
		t.Fatalf("failed to setup clickhouse container: %v", err)
	}
	defer chContainer.Terminate(ctx)

	conn, err := db.InitClickHouse(chContainer.Address)
	if err != nil {
		t.Fatalf("failed to init clickhouse: %v", err)
	}

	err = db.MigrateClickHouseSchema(ctx, conn)
	if err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}

	repo := repository.NewClickHouseAnalyticsRepository(conn)

	// Workspace A and B
	wsA := "wsA-att"
	wsB := "wsB-att"
	now := time.Now().UTC()

	// Insert events representing historical attribution
	batch, err := conn.PrepareBatch(ctx, "INSERT INTO analytics_events")
	if err != nil {
		t.Fatalf("failed to prepare batch: %v", err)
	}

	campA := "camp-a-uuid"
	campB := "camp-b-uuid"
	
	valTwitter := "twitter"
	valGoogle := "google"
	valSocial := "social"
	valCpc := "cpc"
	valSummer := "summer"
	valSummerSale := "summer-sale"

	// Click 1: Link in Campaign A
	_ = batch.Append(
		"evt1", "link.redirect", now.Add(-1*time.Hour), "link1", wsA, "short1", "", "", "ip1",
		&campA, &valTwitter, &valSocial, &valSummer, (*string)(nil), (*string)(nil),
	)

	// Click 2: SAME LINK moved to Campaign B (Historical attribution preservation)
	_ = batch.Append(
		"evt2", "link.redirect", now, "link1", wsA, "short1", "", "", "ip2",
		&campB, &valGoogle, &valCpc, &valSummerSale, (*string)(nil), (*string)(nil),
	)

	// Click 3: Another link in Workspace B to test isolation
	_ = batch.Append(
		"evt3", "link.redirect", now, "linkB", wsB, "shortB", "", "", "ip3",
		&campB, &valGoogle, &valCpc, &valSummerSale, (*string)(nil), (*string)(nil),
	)

	if err := batch.Send(); err != nil {
		t.Fatalf("failed to send batch: %v", err)
	}

	from := now.Add(-24 * time.Hour)
	to := now.Add(24 * time.Hour)

	// --- Campaign Performance Test ---
	campPerf, err := repo.GetCampaignPerformance(ctx, wsA, from, to, 10)
	if err != nil {
		t.Fatalf("failed GetCampaignPerformance: %v", err)
	}
	
	if len(campPerf.Data) != 2 {
		t.Fatalf("expected 2 campaigns in wsA, got %d", len(campPerf.Data))
	}
	
	for _, p := range campPerf.Data {
		if *p.CampaignID == campA {
			if p.Clicks != 1 {
				t.Errorf("expected campA clicks = 1, got %d", p.Clicks)
			}
		} else if *p.CampaignID == campB {
			if p.Clicks != 1 {
				t.Errorf("expected campB clicks = 1, got %d", p.Clicks)
			}
		} else {
			t.Errorf("unexpected campaign: %v", *p.CampaignID)
		}
	}

	// --- UTM Source Performance Test ---
	utmSourcePerf, err := repo.GetUTMPerformance(ctx, wsA, "utm_source", from, to, 10)
	if err != nil {
		t.Fatalf("failed GetUTMPerformance for source: %v", err)
	}
	if len(utmSourcePerf.Data) != 2 {
		t.Fatalf("expected 2 utm sources in wsA, got %d", len(utmSourcePerf.Data))
	}

	for _, p := range utmSourcePerf.Data {
		if p.UTMValue == "twitter" && p.Clicks != 1 {
			t.Errorf("expected twitter clicks = 1, got %d", p.Clicks)
		}
		if p.UTMValue == "google" && p.Clicks != 1 {
			t.Errorf("expected google clicks = 1, got %d", p.Clicks)
		}
	}

	// --- Workspace B Isolation Test ---
	campPerfB, _ := repo.GetCampaignPerformance(ctx, wsB, from, to, 10)
	if len(campPerfB.Data) != 1 {
		t.Fatalf("expected 1 campaign in wsB, got %d", len(campPerfB.Data))
	}
	if *campPerfB.Data[0].CampaignID != campB || campPerfB.Data[0].Clicks != 1 {
		t.Errorf("unexpected data for wsB: %+v", campPerfB.Data[0])
	}
}

func TestClickHouseAnalyticsRepository_GetDomainPerformance(t *testing.T) {
	ctx := context.Background()
	chContainer, err := pkgtesting.SetupClickHouseContainer(ctx)
	require.NoError(t, err)
	defer chContainer.Terminate(ctx)

	conn, err := db.InitClickHouse(chContainer.Address)
	require.NoError(t, err)
	defer conn.Close()

	err = db.MigrateClickHouseSchema(ctx, conn)
	require.NoError(t, err)

	repo := repository.NewClickHouseAnalyticsRepository(conn)
	
	wsID := "ws-domain-test"
	now := time.Now().UTC()
	
	// Create some events for this workspace
	insertQuery := `
		INSERT INTO analytics_events (event_id, event_type, timestamp, link_id, workspace_id, short_code, ip_hash, hostname, custom_domain_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	batch, err := conn.PrepareBatch(ctx, insertQuery)
	require.NoError(t, err)

	batch.Append("e1", "link.redirect", now.Add(-time.Hour), "link1", wsID, "abcd", "ip1", "customer-a.com", "cd-1")
	batch.Append("e2", "link.redirect", now.Add(-time.Hour), "link1", wsID, "abcd", "ip1", "customer-a.com", "cd-1")
	batch.Append("e3", "link.redirect", now.Add(-time.Hour), "link1", wsID, "abcd", "ip2", "customer-a.com", "cd-1")
	
	batch.Append("e4", "link.redirect", now.Add(-time.Hour), "link2", wsID, "efgh", "ip3", "customer-b.com", "cd-2")
	batch.Append("e5", "link.redirect", now.Add(-time.Hour), "link3", wsID, "ijkl", "ip4", nil, nil) // platform domain

	require.NoError(t, batch.Send())
	
	resp, err := repo.GetDomainPerformance(ctx, wsID, now.Add(-2*time.Hour), now, 10)
	require.NoError(t, err)
	
	assert.Len(t, resp.Data, 3)
	
	// Data should be ordered by clicks DESC
	assert.Equal(t, "customer-a.com", resp.Data[0].Hostname)
	assert.Equal(t, uint64(3), resp.Data[0].Clicks)
	assert.Equal(t, uint64(2), resp.Data[0].UniqueVisitors) // ip1, ip2
	
	// Next could be platform or customer-b (both have 1 click)
	// But let's check one of them
	foundPlatform := false
	for _, d := range resp.Data {
		if d.Hostname == "platform" {
			foundPlatform = true
			assert.Equal(t, uint64(1), d.Clicks)
		}
	}
	assert.True(t, foundPlatform)
}
