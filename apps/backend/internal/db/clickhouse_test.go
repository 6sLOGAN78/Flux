package db_test

import (
	"context"
	"testing"
	"time"

	"flux/apps/backend/internal/db"
	pkgtesting "flux/apps/backend/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPhase13_ClickHouseSchemaMigration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	chContainer, err := pkgtesting.SetupClickHouseContainer(ctx)
	require.NoError(t, err, "failed to setup clickhouse container")
	defer chContainer.Terminate(context.Background())

	chConn, err := db.InitClickHouse(chContainer.Address)
	require.NoError(t, err, "failed to init clickhouse")

	// 1. Test Migration Safety (Idempotency)
	err = db.MigrateClickHouseSchema(ctx, chConn)
	require.NoError(t, err, "failed initial schema migration")
	
	err = db.MigrateClickHouseSchema(ctx, chConn)
	require.NoError(t, err, "failed secondary schema migration (idempotency check)")

	// 2. Test Conversions Table Exists & Columns
	var count uint64
	err = chConn.QueryRow(ctx, "SELECT count() FROM system.tables WHERE name = 'conversions'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), count, "conversions table should be created")

	// 3. Test Event ID Bloom Filter Index Exists
	var indexCount uint64
	err = chConn.QueryRow(ctx, "SELECT count() FROM system.data_skipping_indices WHERE table = 'analytics_events' AND name = 'idx_event_id'").Scan(&indexCount)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), indexCount, "idx_event_id bloom filter should exist")

	// 4. Test Conversions Insertion (Workspace Isolation & Array(String) handling)
	batch, err := chConn.PrepareBatch(ctx, "INSERT INTO conversions (conversion_id, workspace_id, timestamp, conversion_name, revenue, currency, click_ids, visitor_id)")
	require.NoError(t, err)

	now := time.Now().UTC()
	err = batch.Append(
		"conv_1",
		"workspace_A",
		now,
		"checkout",
		float64(99.99),
		"USD",
		[]string{"click_1", "click_2"},
		"visitor_x",
	)
	require.NoError(t, err)

	err = batch.Append(
		"conv_2",
		"workspace_B",
		now,
		"signup",
		float64(0.0),
		"USD",
		[]string{"click_3"},
		"visitor_y",
	)
	require.NoError(t, err)

	err = batch.Send()
	require.NoError(t, err, "failed to insert conversion records")

	// 5. Query Validation
	var wsACount uint64
	err = chConn.QueryRow(ctx, "SELECT count() FROM conversions WHERE workspace_id = 'workspace_A'").Scan(&wsACount)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), wsACount, "workspace_A should have 1 conversion")

	var wsBCount uint64
	err = chConn.QueryRow(ctx, "SELECT count() FROM conversions WHERE workspace_id = 'workspace_B'").Scan(&wsBCount)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), wsBCount, "workspace_B should have 1 conversion")

	// 6. Verify click_ids retrieval
	var clickIds []string
	var revenue float64
	err = chConn.QueryRow(ctx, "SELECT click_ids, revenue FROM conversions WHERE conversion_id = 'conv_1'").Scan(&clickIds, &revenue)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"click_1", "click_2"}, clickIds, "click_ids should be stored and retrieved correctly")
	assert.Equal(t, 99.99, revenue, "revenue should match")
}
