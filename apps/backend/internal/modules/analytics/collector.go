// Package analytics provides click event collection, async stream producers, and ingestion handlers.
package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// AsyncCollector collects click events asynchronously without blocking the caller thread.
type AsyncCollector struct {
	producer EventProducer
	eventChan chan *ClickEvent
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewAsyncCollector initializes a new AsyncCollector instance.
func NewAsyncCollector(producer EventProducer, bufferSize int) *AsyncCollector {
	if bufferSize <= 0 {
		bufferSize = 5000
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &AsyncCollector{
		producer:  producer,
		eventChan: make(chan *ClickEvent, bufferSize),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start launches the background worker goroutine loop.
func (c *AsyncCollector) Start(ctx context.Context) {
	c.wg.Add(1)
	go c.workerLoop()
}

// Stop gracefully shuts down the background worker and drains pending events.
func (c *AsyncCollector) Stop() {
	c.cancel()
	close(c.eventChan)
	c.wg.Wait()
}

// CollectAsync pushes a click event to the internal buffered channel non-blockingly.
func (c *AsyncCollector) CollectAsync(event *ClickEvent) {
	if event == nil {
		return
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	select {
	case c.eventChan <- event:
		// Enqueued successfully without blocking
	default:
		// Buffer full under extreme load: log warning and drop event to protect redirect latency budget
		log.Printf("warning: click event buffer full, dropping event for slug '%s'", event.Slug)
	}
}

// workerLoop processes queued click events from the channel and publishes to the producer.
func (c *AsyncCollector) workerLoop() {
	defer c.wg.Done()

	for event := range c.eventChan {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		if err := c.producer.Publish(ctx, event); err != nil {
			log.Printf("error: failed to publish click event for slug '%s': %v", event.Slug, err)
		}
		cancel()
	}
}

// RedisStreamProducer publishes click events to a Redis Stream via XADD.
type RedisStreamProducer struct {
	client     *redis.Client
	streamName string
}

// NewRedisStreamProducer initializes a RedisStreamProducer instance.
func NewRedisStreamProducer(client *redis.Client, streamName string) *RedisStreamProducer {
	if streamName == "" {
		streamName = "stream:click_events"
	}
	return &RedisStreamProducer{
		client:     client,
		streamName: streamName,
	}
}

// Publish writes the ClickEvent payload to the configured Redis Stream using XADD.
func (p *RedisStreamProducer) Publish(ctx context.Context, event *ClickEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal click event: %w", err)
	}

	args := &redis.XAddArgs{
		Stream: p.streamName,
		Values: map[string]interface{}{
			"payload": data,
			"slug":    event.Slug,
			"link_id": event.LinkID,
		},
	}

	return p.client.XAdd(ctx, args).Err()
}
