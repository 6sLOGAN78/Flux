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
}
