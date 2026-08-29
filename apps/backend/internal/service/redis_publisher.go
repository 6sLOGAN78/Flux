package service

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"flux/apps/backend/internal/model/analytics"

	"github.com/redis/go-redis/v9"
)

// RedisAnalyticsPublisher safely publishes events to a Redis Stream using a bounded worker pool.
type RedisAnalyticsPublisher struct {
	client     *redis.Client
	streamName string
	eventChan  chan *analytics.AnalyticsEvent
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewRedisAnalyticsPublisher creates a new publisher with bounded queue and worker.
func NewRedisAnalyticsPublisher(client *redis.Client, streamName string, bufferSize int) *RedisAnalyticsPublisher {
	if streamName == "" {
		streamName = "analytics:events"
	}
	if bufferSize <= 0 {
		bufferSize = 5000 // default bounded queue size
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &RedisAnalyticsPublisher{
		client:     client,
		streamName: streamName,
		eventChan:  make(chan *analytics.AnalyticsEvent, bufferSize),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start launches the background worker goroutine to process XADD non-blockingly.
func (p *RedisAnalyticsPublisher) Start() {
	p.wg.Add(1)
	go p.workerLoop()
}

// Stop gracefully shuts down the background worker and drains pending events up to a timeout.
func (p *RedisAnalyticsPublisher) Stop(timeout time.Duration) {
	p.cancel()
	close(p.eventChan)

	// Wait with timeout
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("RedisAnalyticsPublisher gracefully drained events.")
	case <-time.After(timeout):
		log.Println("RedisAnalyticsPublisher shutdown timed out, dropping remaining events.")
	}
}

// PublishEvent pushes an event to the bounded queue. It does NOT block or spawn goroutines.
func (p *RedisAnalyticsPublisher) PublishEvent(ctx context.Context, event *analytics.AnalyticsEvent) error {
	select {
	case p.eventChan <- event:
		return nil
	default:
		// Queue is full (Redis is likely down or slow). Drop to preserve redirect reliability.
		log.Printf("ERROR: Analytics queue full, dropping event for %s", event.ShortCode)
		return nil
	}
}

func (p *RedisAnalyticsPublisher) workerLoop() {
	defer p.wg.Done()

	for event := range p.eventChan {
		p.publishToRedis(event)
	}
}

func (p *RedisAnalyticsPublisher) publishToRedis(event *analytics.AnalyticsEvent) {
	// Add timeout context for the Redis operation itself
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("ERROR: failed to marshal analytics event: %v", err)
		return
	}

	args := &redis.XAddArgs{
		Stream: p.streamName,
		Values: map[string]interface{}{
			"payload": data,
			"event_id": event.EventID,
		},
	}

	if err := p.client.XAdd(ctx, args).Err(); err != nil {
		// Log the error, but we do not crash or retry indefinitely.
		log.Printf("ERROR: failed to XADD analytics event to redis stream: %v", err)
	}
}
