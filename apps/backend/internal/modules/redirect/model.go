package redirect

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("slug not found")
	ErrExpired  = errors.New("link has expired")
	ErrDisabled = errors.New("link is disabled")
	ErrDeleted  = errors.New("link has been deleted")
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

// LinkRedirectTarget represents the cached redirect instructions for a short link.
type LinkRedirectTarget struct {
	Slug                string     `json:"slug"`
	DestinationURL      string     `json:"destination_url"`
	Status              string     `json:"status"` // "active", "disabled", "expired", "deleted"
	ExpiresAt           *time.Time `json:"expires_at,omitempty"`
	IsPasswordProtected bool       `json:"is_password_protected"`
	RedirectCode        int        `json:"redirect_code"` // 301 or 302
}

// RedirectRepository defines persistence methods for short links.
type RedirectRepository interface {
	GetBySlug(ctx context.Context, slug string) (*LinkRedirectTarget, error)
}

// RedirectCache defines caching operations for redirect targets.
type RedirectCache interface {
	Get(ctx context.Context, slug string) (*LinkRedirectTarget, error)
	Set(ctx context.Context, slug string, target *LinkRedirectTarget, ttl time.Duration) error
	Delete(ctx context.Context, slug string) error
}
