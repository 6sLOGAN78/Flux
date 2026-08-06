package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

// LimiterStore defines the contract for sliding window rate limit evaluation.
type LimiterStore interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (allowed bool, remaining int, resetIn time.Duration, err error)
}

// RedisSlidingWindowLimiter implements LimiterStore using Redis Sorted Sets (ZSET).
type RedisSlidingWindowLimiter struct {
	client *redis.Client
}

// NewRedisSlidingWindowLimiter initializes a Redis-backed sliding window rate limiter.
func NewRedisSlidingWindowLimiter(client *redis.Client) *RedisSlidingWindowLimiter {
	return &RedisSlidingWindowLimiter{client: client}
}

// Allow evaluates a request against a sliding time window quota in Redis.
func (r *RedisSlidingWindowLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, int, time.Duration, error) {
	if r.client == nil {
		return true, limit, 0, nil
	}

	redisKey := fmt.Sprintf("ratelimit:%s", key)
	now := time.Now()
	nowMicro := now.UnixNano() / 1000
	windowStartMicro := now.Add(-window).UnixNano() / 1000

	pipe := r.client.Pipeline()
	pipe.ZRemRangeByScore(ctx, redisKey, "0", strconv.FormatInt(windowStartMicro, 10))
	cardCmd := pipe.ZCard(ctx, redisKey)
	pipe.ZAdd(ctx, redisKey, redis.Z{Score: float64(nowMicro), Member: nowMicro})
	pipe.Expire(ctx, redisKey, window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return true, limit, 0, err
	}

	count := int(cardCmd.Val())
	if count >= limit {
		return false, 0, window, nil
	}

	remaining := limit - (count + 1)
	return true, remaining, window, nil
}

// RateLimitMiddleware creates an Echo middleware enforcing request quotas.
func RateLimitMiddleware(store LimiterStore, limit int, window time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := c.RealIP()
			if apiKey := c.Request().Header.Get("X-API-Key"); apiKey != "" {
				key = "key:" + apiKey
			} else if userID, ok := c.Get("user_id").(string); ok && userID != "" {
				key = "user:" + userID
			}

			allowed, remaining, resetIn, err := store.Allow(c.Request().Context(), key, limit, window)
			if err != nil {
				// Soft fallback: allow request if Redis fails to avoid blocking legitimate traffic
				return next(c)
			}

			c.Response().Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
			c.Response().Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			c.Response().Header().Set("X-RateLimit-Reset", strconv.FormatInt(int64(resetIn.Seconds()), 10))

			if !allowed {
				return echo.NewHTTPError(http.StatusTooManyRequests, "rate limit quota exceeded")
			}

			return next(c)
		}
	}
}
