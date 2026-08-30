package redirect

import (
	"time"
)

// ReservedSlugs contains system routes that should never be processed as short link slugs.
var ReservedSlugs = map[string]bool{
	"login":       true,
	"signup":      true,
	"api":         true,
	"docs":        true,
	"health":      true,
	"static":      true,
	"favicon.ico": true,
	"robots.txt":  true,
	"dashboard":   true,
}

// LinkRedirectTarget represents the resolved redirect target for a short link.
type LinkRedirectTarget struct {
	Slug                string     `json:"slug"`
	LinkID              string     `json:"link_id"`
	TenantID            string     `json:"tenant_id"`
	DestinationURL      string     `json:"destination_url"`
	Status              string     `json:"status"` // "active", "disabled", "expired", "deleted"
	ExpiresAt           *time.Time `json:"expires_at,omitempty"`
	IsPasswordProtected bool       `json:"is_password_protected"`
	RedirectCode        int        `json:"redirect_code"` // 301 or 302
	CampaignID          *string    `json:"campaign_id,omitempty"`
	UTMSource           *string    `json:"utm_source,omitempty"`
	UTMMedium           *string    `json:"utm_medium,omitempty"`
	UTMCampaign         *string    `json:"utm_campaign,omitempty"`
	UTMTerm             *string    `json:"utm_term,omitempty"`
	UTMContent          *string    `json:"utm_content,omitempty"`
	Hostname            string     `json:"hostname"`
	CustomDomainID      *string    `json:"custom_domain_id,omitempty"`
}
