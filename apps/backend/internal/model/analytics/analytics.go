package analytics

import (
	"time"
)

// ClickEvent represents an un-aggregated click event payload.
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

// AnalyticsSummaryResponse represents aggregated user dashboard statistics.
type AnalyticsSummaryResponse struct {
	TotalClicks  int64            `json:"total_clicks"`
	UniqueUsers  int64            `json:"unique_users"`
	TopCountries map[string]int64 `json:"top_countries"`
	TopBrowsers  map[string]int64 `json:"top_browsers"`
	TopDevices   map[string]int64 `json:"top_devices"`
	Page         int              `json:"page"`
	Limit        int              `json:"limit"`
}

// LinkMetricsResponse represents analytics breakdown for a single short link.
type LinkMetricsResponse struct {
	LinkID      string           `json:"link_id"`
	TotalClicks int64            `json:"total_clicks"`
	DailyStats  map[string]int64 `json:"daily_stats"`
}
