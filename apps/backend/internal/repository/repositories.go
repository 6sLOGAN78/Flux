// Package repository defines data access interfaces and persistence implementations.
package repository

import (
	"context"
	"errors"
	"time"

	"flux/apps/backend/internal/model/analytics"
	"flux/apps/backend/internal/model/redirect"
)

var (
	ErrNotFound = errors.New("resource not found")
)

// RedirectRepository defines link persistence operations.
type RedirectRepository interface {
	GetBySlug(ctx context.Context, slug string) (*redirect.LinkRedirectTarget, error)
}

// RedirectCache defines Redis caching operations for redirect targets.
type RedirectCache interface {
	Get(ctx context.Context, slug string) (*redirect.LinkRedirectTarget, error)
	Set(ctx context.Context, slug string, target *redirect.LinkRedirectTarget, ttl time.Duration) error
	Delete(ctx context.Context, slug string) error
}

// EventProducer defines publishing operations for click events to Redis Stream / Kafka.
type EventProducer interface {
	Publish(ctx context.Context, event *analytics.ClickEvent) error
}

// AnalyticsProvider defines querying operations for dashboard metrics.
type AnalyticsProvider interface {
	GetSummary(ctx context.Context, userID string, page, limit int) (*analytics.AnalyticsSummaryResponse, error)
	GetLinkMetrics(ctx context.Context, linkID string) (*analytics.LinkMetricsResponse, error)
}

// QRCache defines caching operations for generated QR images.
type QRCache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, data []byte) error
}
