package repository

import (
	"context"

	"github.com/google/uuid"
)

// GetHostnameForLink efficiently retrieves the custom domain hostname associated with a link, if any.
func (r *LinkRepository) GetHostnameForLink(ctx context.Context, linkID uuid.UUID) (*string, error) {
	stmt := `
		SELECT cd.hostname
		FROM links l
		JOIN custom_domains cd ON l.custom_domain_id = cd.id
		WHERE l.id = $1
	`
	var hostname string
	err := r.pool.QueryRow(ctx, stmt, linkID).Scan(&hostname)
	if err != nil {
		// If ErrNoRows, it just means no custom domain is attached. Return nil.
		return nil, nil
	}
	return &hostname, nil
}
