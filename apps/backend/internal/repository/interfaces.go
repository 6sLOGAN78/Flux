package repository

import (
	"context"
	"errors"
	"time"

	"flux/apps/backend/internal/model/analytics"
	"flux/apps/backend/internal/model/redirect"
)

var ErrNotFound = errors.New("repository: resource not found")

type RedirectRepository interface {
	GetByHostAndSlug(ctx context.Context, hostname, slug string) (*redirect.LinkRedirectTarget, error)
}

type RedirectCache interface {
	Get(ctx context.Context, hostname, slug string) (*redirect.LinkRedirectTarget, error)
	Set(ctx context.Context, hostname, slug string, target *redirect.LinkRedirectTarget, ttl time.Duration) error
	Delete(ctx context.Context, hostname, slug string) error
	DeleteHost(ctx context.Context, hostname string) error
}

type QRCache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, data []byte) error
}

type AnalyticsProvider interface {
	GetSummary(ctx context.Context, workspaceID string, from, to time.Time) (*analytics.AnalyticsSummary, error)
	GetTimeseries(ctx context.Context, workspaceID string, from, to time.Time, interval string) (*analytics.TimeseriesResponse, error)
	GetTopLinks(ctx context.Context, workspaceID string, from, to time.Time, limit int) (*analytics.TopLinksResponse, error)
	GetReferrers(ctx context.Context, workspaceID string, from, to time.Time, limit int) (*analytics.ReferrersResponse, error)
	GetCampaignPerformance(ctx context.Context, workspaceID string, from, to time.Time, limit int) (*analytics.CampaignPerformanceResponse, error)
	GetUTMPerformance(ctx context.Context, workspaceID string, dimension string, from, to time.Time, limit int) (*analytics.UTMPerformanceResponse, error)
}
