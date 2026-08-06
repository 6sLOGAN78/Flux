package redirect

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisRedirectCache implements RedirectCache using Redis v9.
type RedisRedirectCache struct {
	client *redis.Client
}

// NewRedisRedirectCache initializes a Redis-backed redirect cache.
func NewRedisRedirectCache(client *redis.Client) *RedisRedirectCache {
	return &RedisRedirectCache{client: client}
}

func (r *RedisRedirectCache) Get(ctx context.Context, slug string) (*LinkRedirectTarget, error) {
	key := fmt.Sprintf("link:%s", slug)
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var target LinkRedirectTarget
	if err := json.Unmarshal([]byte(val), &target); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached link: %w", err)
	}

	return &target, nil
}

func (r *RedisRedirectCache) Set(ctx context.Context, slug string, target *LinkRedirectTarget, ttl time.Duration) error {
	key := fmt.Sprintf("link:%s", slug)
	data, err := json.Marshal(target)
	if err != nil {
		return fmt.Errorf("failed to marshal link for cache: %w", err)
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisRedirectCache) Delete(ctx context.Context, slug string) error {
	key := fmt.Sprintf("link:%s", slug)
	return r.client.Del(ctx, key).Err()
}
