package analytics

import (
	"context"
	"time"
)

// EventType defines the type of analytics event.
type EventType string

const (
	EventTypeLinkRedirect EventType = "link.redirect"
)

// AnalyticsEvent represents the canonical analytics event.
type AnalyticsEvent struct {
	EventID     string    `json:"event_id"`
	EventType   EventType `json:"event_type"`
	Timestamp   time.Time `json:"timestamp"`
	LinkID      string    `json:"link_id"`
	WorkspaceID string    `json:"workspace_id"`
	ShortCode   string    `json:"short_code"`

	// Metadata
	Referrer  string `json:"referrer,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
	
	// Privacy-Preserving
	IPHash string `json:"ip_hash,omitempty"` 

	// Enriched Data (Future)
	Country    string `json:"country,omitempty"`
	City       string `json:"city,omitempty"`
	Browser    string `json:"browser,omitempty"`
	OS         string `json:"os,omitempty"`
	DeviceType string `json:"device_type,omitempty"`
}

// AnalyticsPublisher defines the abstraction for publishing events asynchronously.
type AnalyticsPublisher interface {
	PublishEvent(ctx context.Context, event *AnalyticsEvent) error
}

// API Data Transfer Objects

type AnalyticsSummary struct {
	TotalClicks    uint64 `json:"total_clicks"`
	UniqueVisitors uint64 `json:"unique_visitors"`
}

type TimeseriesDataPoint struct {
	Timestamp      string `json:"timestamp"`
	Clicks         uint64 `json:"clicks"`
	UniqueVisitors uint64 `json:"unique_visitors"`
}

type TimeseriesResponse struct {
	Data []TimeseriesDataPoint `json:"data"`
}

type TopLink struct {
	LinkID    string `json:"link_id"`
	ShortCode string `json:"short_code"`
	Clicks    uint64 `json:"clicks"`
}

type TopLinksResponse struct {
	Data []TopLink `json:"data"`
}

type ReferrerStat struct {
	Referrer string `json:"referrer"`
	Clicks   uint64 `json:"clicks"`
}

type ReferrersResponse struct {
	Data []ReferrerStat `json:"data"`
}
