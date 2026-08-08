package repository

import (
	"context"
	"errors"
	"time"

	"flux/apps/backend/internal/model/analytics"
	"flux/apps/backend/internal/model/redirect"
)

// ErrNotFound indicates requested entity was not found in storage.
var ErrNotFound = errors.New("repository: resource not found")

// RedirectRepository defines data access for redirect targets.
type RedirectRepository interface {
	GetBySlug(ctx context.Context, slug string) (*redirect.LinkRedirectTarget, error)
}

// RedirectCache defines caching operations for link redirects.
type RedirectCache interface {
	Get(ctx context.Context, slug string) (*redirect.LinkRedirectTarget, error)
	Set(ctx context.Context, slug string, target *redirect.LinkRedirectTarget, ttl time.Duration) error
	Delete(ctx context.Context, slug string) error
}

// EventProducer defines publishing for analytics events.
type EventProducer interface {
	Publish(ctx context.Context, event *analytics.ClickEvent) error
}

// QRCache defines caching for generated QR code binaries.
type QRCache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, data []byte) error
}

// AnalyticsProvider defines read operations for analytics summaries and link metrics.
type AnalyticsProvider interface {
	GetSummary(ctx context.Context, userID string, page, limit int) (any, error)
	GetLinkMetrics(ctx context.Context, linkID string) (any, error)
}

