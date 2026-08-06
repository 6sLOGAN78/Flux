package analytics

import (
	"context"
	"time"
)

// ClickEvent represents a raw, un-aggregated click event payload.
type ClickEvent struct {
	ID          string    `json:"id"`
	LinkID      string    `json:"link_id"`
	Slug        string    `json:"slug"`
	Timestamp   time.Time `json:"timestamp"`
	IPAddress   string    `json:"ip_address"`
	UserAgent   string    `json:"user_agent"`
	Referrer    string    `json:"referrer"`
	Country     string    `json:"country,omitempty"`
	City        string    `json:"city,omitempty"`
	Browser     string    `json:"browser,omitempty"`
	OS          string    `json:"os,omitempty"`
	DeviceType  string    `json:"device_type,omitempty"`
}

// EventProducer defines the interface for pushing click events to an async message queue (Redis Stream / Kafka).
type EventProducer interface {
	Publish(ctx context.Context, event *ClickEvent) error
}
