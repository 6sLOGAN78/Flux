package analytics_test

import (
	"context"
	"testing"
	"time"

	"flux/apps/backend/internal/modules/analytics"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGlobalAnalytics_BatchAndCompressEvents(t *testing.T) {
	batcher := analytics.NewEdgeBatcher(10, 500*time.Millisecond)

	event := analytics.ClickEvent{
		ID:             uuid.New(),
		LinkID:         uuid.New(),
		UserID:         uuid.New(),
		Timestamp:      time.Now(),
		IPAddress:      "1.2.3.4",
		CountryCode:    "US",
		UserAgent:      "Mozilla/5.0",
		DeviceType:     "desktop",
		ReferrerDomain: "google.com",
	}

	ctx := context.Background()
	err := batcher.AddEvent(ctx, event)
	require.NoError(t, err)

	assert.Equal(t, 1, batcher.PendingCount())

	batch, err := batcher.Flush(ctx)
	require.NoError(t, err)
	require.NotNil(t, batch)
	assert.Len(t, batch.Events, 1)
	assert.NotEmpty(t, batch.CompressedData)

	decompressedEvents, err := batcher.DecompressBatch(batch.CompressedData)
	require.NoError(t, err)
	require.Len(t, decompressedEvents, 1)
	assert.Equal(t, event.ID, decompressedEvents[0].ID)
	assert.Equal(t, event.IPAddress, decompressedEvents[0].IPAddress)
}

func TestGlobalAnalytics_IngestCompressedStream(t *testing.T) {
	batcher := analytics.NewEdgeBatcher(5, 500*time.Millisecond)
	mockStore := analytics.NewMockClickHouseStore()
	ingestor := analytics.NewGlobalStreamIngestor(mockStore)

	ctx := context.Background()
	for i := 0; i < 3; i++ {
		err := batcher.AddEvent(ctx, analytics.ClickEvent{
			ID:          uuid.New(),
			LinkID:      uuid.New(),
			UserID:      uuid.New(),
			Timestamp:   time.Now(),
			IPAddress:   "5.6.7.8",
			CountryCode: "DE",
		})
		require.NoError(t, err)
	}

	batch, err := batcher.Flush(ctx)
	require.NoError(t, err)

	count, err := ingestor.IngestCompressedStream(ctx, batch.CompressedData)
	require.NoError(t, err)
	assert.Equal(t, 3, count)

	metrics, err := ingestor.GetStreamMetrics(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(3), metrics.TotalIngestedEvents)
	assert.True(t, metrics.CompressionRatio > 0)
}
