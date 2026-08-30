package repository

import (
	"context"

	"github.com/google/uuid"
)

type LinkCacheKey struct {
	Hostname string
	Slug     string
}

func (r *LinkRepository) GetCacheKeysByCampaign(ctx context.Context, campaignID uuid.UUID) ([]LinkCacheKey, error) {
	stmt := `
		SELECT l.short_code, COALESCE(cd.hostname, '') as hostname
		FROM links l
		LEFT JOIN custom_domains cd ON l.custom_domain_id = cd.id
		WHERE l.campaign_id = $1
	`
	rows, err := r.pool.Query(ctx, stmt, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []LinkCacheKey
	for rows.Next() {
		var key LinkCacheKey
		if err := rows.Scan(&key.Slug, &key.Hostname); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}

	return keys, nil
}
