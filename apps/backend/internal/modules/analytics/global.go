// Package analytics provides global high-throughput click event stream batching, snappy/gzip compression, and multi-region ingestion.
package analytics

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// ClickEventBatch represents a aggregated chunk of click events ready for compressed edge streaming.
type ClickEventBatch struct {
	BatchID        string       `json:"batch_id"`
	Timestamp      time.Time    `json:"timestamp"`
	Events         []ClickEvent `json:"events"`
	CompressedData []byte       `json:"-"`
}

// EdgeBatcher buffers incoming edge click events and flushes compressed batches to regional ingestion brokers.
type EdgeBatcher struct {
	batchSize     int
	flushInterval time.Duration
	pending       []ClickEvent
	mu            sync.Mutex
}

// NewEdgeBatcher initializes an EdgeBatcher with specified capacity limit and flush interval.
func NewEdgeBatcher(batchSize int, flushInterval time.Duration) *EdgeBatcher {
	return &EdgeBatcher{
		batchSize:     batchSize,
		flushInterval: flushInterval,
		pending:       make([]ClickEvent, 0, batchSize),
	}
}

// AddEvent appends a click event to the edge buffer.
func (b *EdgeBatcher) AddEvent(ctx context.Context, event ClickEvent) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.pending = append(b.pending, event)
	return nil
}

// PendingCount returns the number of un-flushed events currently in the buffer.
func (b *EdgeBatcher) PendingCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.pending)
}

// Flush packages buffered events into a compressed ClickEventBatch and resets the internal buffer.
func (b *EdgeBatcher) Flush(ctx context.Context) (*ClickEventBatch, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.pending) == 0 {
		return nil, nil
	}

	events := make([]ClickEvent, len(b.pending))
	copy(events, b.pending)
	b.pending = b.pending[:0]

	batch := &ClickEventBatch{
		BatchID:   uuid.New().String(),
		Timestamp: time.Now().UTC(),
		Events:    events,
	}

	compressed, err := b.CompressBatch(batch)
	if err != nil {
		return nil, fmt.Errorf("failed to compress batch: %w", err)
	}
	batch.CompressedData = compressed
	return batch, nil
}

// CompressBatch serializes and gzip-compresses event batch data into binary wire payload.
func (b *EdgeBatcher) CompressBatch(batch *ClickEventBatch) ([]byte, error) {
	data, err := json.Marshal(batch.Events)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(data); err != nil {
		_ = gw.Close()
		return nil, err
	}
	if err := gw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DecompressBatch inflates compressed wire payload back into slice of ClickEvent objects.
func (b *EdgeBatcher) DecompressBatch(compressed []byte) ([]ClickEvent, error) {
	gr, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gr.Close()

	decompressed, err := io.ReadAll(gr)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress gzip stream: %w", err)
	}

	var events []ClickEvent
	if err := json.Unmarshal(decompressed, &events); err != nil {
		return nil, fmt.Errorf("failed to unmarshal click events: %w", err)
	}
	return events, nil
}

// StreamMetrics tracks global analytics ingestion pipeline metrics.
type StreamMetrics struct {
	TotalIngestedEvents  int64   `json:"total_ingested_events"`
	TotalBytesProcessed  int64   `json:"total_bytes_processed"`
	CompressionRatio     float64 `json:"compression_ratio"`
}

// GlobalStreamIngestor consumes compressed click streams and writes events to ClickHouse storage tables.
type GlobalStreamIngestor struct {
	store                ClickHouseStore
	totalEvents          int64
	totalRawBytes        int64
	totalCompressedBytes int64
}

// NewGlobalStreamIngestor constructs a GlobalStreamIngestor backed by ClickHouse storage.
func NewGlobalStreamIngestor(store ClickHouseStore) *GlobalStreamIngestor {
	return &GlobalStreamIngestor{
		store: store,
	}
}

// IngestCompressedStream inflates incoming compressed stream payloads and inserts events into ClickHouse.
func (i *GlobalStreamIngestor) IngestCompressedStream(ctx context.Context, compressedPayload []byte) (int, error) {
	if len(compressedPayload) == 0 {
		return 0, nil
	}

	gr, err := gzip.NewReader(bytes.NewReader(compressedPayload))
	if err != nil {
		return 0, fmt.Errorf("invalid compressed payload: %w", err)
	}
	defer gr.Close()

	rawBytes, err := io.ReadAll(gr)
	if err != nil {
		return 0, fmt.Errorf("failed to decompress payload: %w", err)
	}

	var events []ClickEvent
	if err := json.Unmarshal(rawBytes, &events); err != nil {
		return 0, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if i.store != nil {
		if err := i.store.InsertBatch(ctx, events); err != nil {
			return 0, fmt.Errorf("failed to insert batch to clickhouse: %w", err)
		}
	}

	count := int64(len(events))
	atomic.AddInt64(&i.totalEvents, count)
	atomic.AddInt64(&i.totalRawBytes, int64(len(rawBytes)))
	atomic.AddInt64(&i.totalCompressedBytes, int64(len(compressedPayload)))

	return len(events), nil
}

// GetStreamMetrics retrieves current stream ingestion counters and compression efficiency ratio.
func (i *GlobalStreamIngestor) GetStreamMetrics(ctx context.Context) (*StreamMetrics, error) {
	raw := atomic.LoadInt64(&i.totalRawBytes)
	comp := atomic.LoadInt64(&i.totalCompressedBytes)

	var ratio float64 = 1.0
	if comp > 0 {
		ratio = float64(raw) / float64(comp)
	}

	return &StreamMetrics{
		TotalIngestedEvents: atomic.LoadInt64(&i.totalEvents),
		TotalBytesProcessed: raw,
		CompressionRatio:    ratio,
	}, nil
}
