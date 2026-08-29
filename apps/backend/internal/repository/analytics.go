package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"flux/apps/backend/internal/model/analytics"

	"github.com/redis/go-redis/v9"
)

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
func (p *RedisStreamProducer) Publish(ctx context.Context, event *analytics.AnalyticsEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal click event: %w", err)
	}

	args := &redis.XAddArgs{
		Stream: p.streamName,
		Values: map[string]interface{}{
			"payload": data,
			"slug":    event.ShortCode,
			"link_id": event.LinkID,
		},
	}

	return p.client.XAdd(ctx, args).Err()
}
