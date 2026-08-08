// Package analytics handles time-series click event batching and ClickHouse ReplacingMergeTree persistence.
package analytics

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ClickEvent represents a time-series click event fact.
type ClickEvent struct {
	ID             uuid.UUID  `json:"id" ch:"id"`
	LinkID         uuid.UUID  `json:"link_id" ch:"link_id"`
	DomainID       *uuid.UUID `json:"domain_id,omitempty" ch:"domain_id"`
	UserID         uuid.UUID  `json:"user_id" ch:"user_id"`
	Timestamp      time.Time  `json:"timestamp" ch:"timestamp"`
	IPAddress      string     `json:"ip_address" ch:"ip_address"`
	CountryCode    string     `json:"country_code" ch:"country_code"`
	Region         string     `json:"region" ch:"region"`
	City           string     `json:"city" ch:"city"`
	Latitude       *float64   `json:"latitude,omitempty" ch:"latitude"`
	Longitude      *float64   `json:"longitude,omitempty" ch:"longitude"`
	UserAgent      string     `json:"user_agent" ch:"user_agent"`
	Browser        string     `json:"browser" ch:"browser"`
	BrowserVersion string     `json:"browser_version" ch:"browser_version"`
	OS             string     `json:"os" ch:"os"`
	OSVersion      string     `json:"os_version" ch:"os_version"`
	DeviceType     string     `json:"device_type" ch:"device_type"`
	ReferrerDomain string     `json:"referrer_domain" ch:"referrer_domain"`
	ReferrerURL    string     `json:"referrer_url" ch:"referrer_url"`
	UTMSource      string     `json:"utm_source" ch:"utm_source"`
	UTMMedium      string     `json:"utm_medium" ch:"utm_medium"`
	UTMCampaign    string     `json:"utm_campaign" ch:"utm_campaign"`
	UTMTerm        string     `json:"utm_term" ch:"utm_term"`
	UTMContent     string     `json:"utm_content" ch:"utm_content"`
	QRCodeScan     uint8      `json:"qr_code_scan" ch:"qr_code_scan"`
	ResponseTimeMS uint32     `json:"response_time_ms" ch:"response_time_ms"`
}

// TimeSeriesPoint represents aggregated time-series metric data.
type TimeSeriesPoint struct {
	Timestamp      time.Time `json:"timestamp"`
	Clicks         int64     `json:"clicks"`
	UniqueVisitors int64     `json:"unique_visitors"`
}

// ClickHouseStore defines the storage interface for ClickHouse operations.
type ClickHouseStore interface {
	InsertBatch(ctx context.Context, events []ClickEvent) error
	QueryTimeSeries(ctx context.Context, linkID string, from, to time.Time, granularity string) ([]TimeSeriesPoint, error)
}

// MockClickHouseStore provides an in-memory ClickHouseStore implementation for unit testing.
type MockClickHouseStore struct {
	mu     sync.Mutex
	events []ClickEvent
}

func NewMockClickHouseStore() *MockClickHouseStore {
	return &MockClickHouseStore{
		events: make([]ClickEvent, 0),
	}
}

func (m *MockClickHouseStore) InsertBatch(ctx context.Context, events []ClickEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, events...)
	return nil
}

func (m *MockClickHouseStore) AddEvents(events []ClickEvent) {
	_ = m.InsertBatch(context.Background(), events)
}

func (m *MockClickHouseStore) GetEvents() []ClickEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]ClickEvent, len(m.events))
	copy(result, m.events)
	return result
}

func (m *MockClickHouseStore) QueryTimeSeries(ctx context.Context, linkID string, from, to time.Time, granularity string) ([]TimeSeriesPoint, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	parsedID, err := uuid.Parse(linkID)
	if err != nil {
		return nil, fmt.Errorf("invalid link id: %w", err)
	}

	ipSet := make(map[string]bool)
	var count int64

	for _, e := range m.events {
		if e.LinkID == parsedID && !e.Timestamp.Before(from) && !e.Timestamp.After(to) {
			count++
			if e.IPAddress != "" {
				ipSet[e.IPAddress] = true
			}
		}
	}

	return []TimeSeriesPoint{
		{
			Timestamp:      from,
			Clicks:         count,
			UniqueVisitors: int64(len(ipSet)),
		},
	}, nil
}

// ClickHouseBatchWriter buffers and writes click events asynchronously in batches.
type ClickHouseBatchWriter struct {
	store         ClickHouseStore
	batchSize     int
	flushInterval time.Duration
	eventChan     chan ClickEvent
	mu            sync.Mutex
	buffer        []ClickEvent
}

// NewClickHouseBatchWriter initializes a ClickHouseBatchWriter instance.
func NewClickHouseBatchWriter(store ClickHouseStore, batchSize int, flushInterval time.Duration) *ClickHouseBatchWriter {
	if batchSize <= 0 {
		batchSize = 1000
	}
	if flushInterval <= 0 {
		flushInterval = 1 * time.Second
	}
	return &ClickHouseBatchWriter{
		store:         store,
		batchSize:     batchSize,
		flushInterval: flushInterval,
		eventChan:     make(chan ClickEvent, batchSize*2),
		buffer:        make([]ClickEvent, 0, batchSize),
	}
}

// Start launches the asynchronous background worker to process and flush event batches.
func (w *ClickHouseBatchWriter) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(w.flushInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				_ = w.Flush(context.Background())
				return
			case event, ok := <-w.eventChan:
				if !ok {
					_ = w.Flush(context.Background())
					return
				}
				w.mu.Lock()
				w.buffer = append(w.buffer, event)
				shouldFlush := len(w.buffer) >= w.batchSize
				w.mu.Unlock()

				if shouldFlush {
					_ = w.Flush(ctx)
				}
			case <-ticker.C:
				_ = w.Flush(ctx)
			}
		}
	}()
}

// Add queues a click event for batch writing.
func (w *ClickHouseBatchWriter) Add(event ClickEvent) {
	select {
	case w.eventChan <- event:
	default:
		w.mu.Lock()
		w.buffer = append(w.buffer, event)
		w.mu.Unlock()
	}
}

// Flush flushes buffered click events directly to ClickHouse storage.
func (w *ClickHouseBatchWriter) Flush(ctx context.Context) error {
	w.mu.Lock()
	if len(w.buffer) == 0 {
		w.mu.Unlock()
		return nil
	}
	toInsert := w.buffer
	w.buffer = make([]ClickEvent, 0, w.batchSize)
	w.mu.Unlock()

	if w.store == nil {
		return nil
	}

	return w.store.InsertBatch(ctx, toInsert)
}
