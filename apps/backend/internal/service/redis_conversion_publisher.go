package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"flux/apps/backend/internal/model/analytics"

	"github.com/redis/go-redis/v9"
)

type RedisConversionPublisher struct {
	client     *redis.Client
	streamName string
}

func NewRedisConversionPublisher(client *redis.Client, streamName string) *RedisConversionPublisher {
	if streamName == "" {
		streamName = "analytics:conversions"
	}
	return &RedisConversionPublisher{
		client:     client,
		streamName: streamName,
	}
}

func (p *RedisConversionPublisher) PublishConversion(ctx context.Context, event *analytics.ConversionEvent) error {
	// Directly XAdd to Redis with a short timeout.
	// We do not queue these in memory because we want to provide synchronous 
	// backpressure to the HTTP caller if Redis is down, rather than dropping them silently.
	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal conversion: %w", err)
	}

	args := &redis.XAddArgs{
		Stream: p.streamName,
		Values: map[string]interface{}{
			"payload": data,
			"conversion_id": event.ConversionID,
		},
	}

	if err := p.client.XAdd(timeoutCtx, args).Err(); err != nil {
		return fmt.Errorf("xadd conversion to redis: %w", err)
	}

	return nil
}
