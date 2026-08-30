package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"flux/apps/backend/internal/model/redirect"
	"flux/apps/backend/pkg/sqlerr"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// RedisRedirectCache implements RedirectCache using Redis v9.
type RedisRedirectCache struct {
	client *redis.Client
}

// NewRedisRedirectCache initializes a Redis-backed redirect cache.
func NewRedisRedirectCache(client *redis.Client) *RedisRedirectCache {
	return &RedisRedirectCache{client: client}
}

func (r *RedisRedirectCache) Get(ctx context.Context, hostname, slug string) (*redirect.LinkRedirectTarget, error) {
	key := fmt.Sprintf("redirect:%s:%s", hostname, slug)
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var target redirect.LinkRedirectTarget
	if err := json.Unmarshal([]byte(val), &target); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached link: %w", err)
	}

	return &target, nil
}

func (r *RedisRedirectCache) Set(ctx context.Context, hostname, slug string, target *redirect.LinkRedirectTarget, ttl time.Duration) error {
	key := fmt.Sprintf("redirect:%s:%s", hostname, slug)
	data, err := json.Marshal(target)
	if err != nil {
		return fmt.Errorf("failed to marshal link for cache: %w", err)
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisRedirectCache) Delete(ctx context.Context, hostname, slug string) error {
	key := fmt.Sprintf("redirect:%s:%s", hostname, slug)
	return r.client.Del(ctx, key).Err()
}

func (r *RedisRedirectCache) DeleteHost(ctx context.Context, hostname string) error {
	var cursor uint64
	match := fmt.Sprintf("redirect:%s:*", hostname)
	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, match, 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := r.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}

type PostgresRedirectRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRedirectRepository(pool *pgxpool.Pool) *PostgresRedirectRepository {
	return &PostgresRedirectRepository{pool: pool}
}

func (r *PostgresRedirectRepository) GetByHostAndSlug(ctx context.Context, hostname, slug string) (*redirect.LinkRedirectTarget, error) {
	stmt := `
		SELECT 
			l.id::text, 
			l.tenant_id::text, 
			l.destination_url,
			l.campaign_id::text as campaign_id,
			l.utm_source,
			l.utm_medium,
			l.utm_campaign as link_utm_campaign,
			l.utm_term,
			l.utm_content,
			c.utm_campaign as camp_utm_campaign,
			cd.hostname as custom_hostname,
			cd.status as custom_domain_status,
			l.custom_domain_id::text as custom_domain_id,
			(SELECT tenant_id FROM custom_domains WHERE hostname = @hostname LIMIT 1)::text as req_host_tenant_id,
			(SELECT status FROM custom_domains WHERE hostname = @hostname LIMIT 1) as req_host_status
		FROM links l
		LEFT JOIN campaigns c ON l.campaign_id = c.id
		LEFT JOIN custom_domains cd ON l.custom_domain_id = cd.id
		WHERE l.short_code = @slug
	`
	rows, err := r.pool.Query(ctx, stmt, pgx.NamedArgs{"slug": slug, "hostname": hostname})
	if err != nil {
		return nil, sqlerr.HandleError(fmt.Errorf("failed to query redirect: %w", err))
	}
	
	type rowType struct {
		ID                 string  `db:"id"`
		TenantID           string  `db:"tenant_id"`
		DestinationURL     string  `db:"destination_url"`
		CampaignID         *string `db:"campaign_id"`
		UTMSource          *string `db:"utm_source"`
		UTMMedium          *string `db:"utm_medium"`
		LinkUTMCampaign    *string `db:"link_utm_campaign"`
		UTMTerm            *string `db:"utm_term"`
		UTMContent         *string `db:"utm_content"`
		CampUTMCampaign    *string `db:"camp_utm_campaign"`
		CustomHostname     *string `db:"custom_hostname"`
		CustomDomainStatus *string `db:"custom_domain_status"`
		CustomDomainID     *string `db:"custom_domain_id"`
		ReqHostTenantID    *string `db:"req_host_tenant_id"`
		ReqHostStatus      *string `db:"req_host_status"`
	}
	
	item, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[rowType])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, sqlerr.HandleError(fmt.Errorf("failed to collect redirect row: %w", err))
	}

	// Hostname authorization/routing check
	if item.ReqHostTenantID != nil {
		// 1. Requested hostname is a registered custom domain. Must be ACTIVE.
		if *item.ReqHostStatus != "active" {
			return nil, ErrNotFound
		}
		// 2. Cross-tenant leakage blocked.
		if item.TenantID != *item.ReqHostTenantID {
			return nil, ErrNotFound
		}
		// 3. If link is bound strictly to a different custom domain.
		if item.CustomDomainID != nil && item.CustomHostname != nil && *item.CustomHostname != hostname {
			return nil, ErrNotFound
		}
	} else {
		// Requested hostname is NOT a registered custom domain (assumed platform domain).
		// Prevent serving links strictly bound to a custom domain.
		if item.CustomDomainID != nil {
			return nil, ErrNotFound
		}
	}

	// Resolve UTM Campaign (Link override wins over Campaign default)
	resolvedUTMCampaign := item.LinkUTMCampaign
	if resolvedUTMCampaign == nil {
		resolvedUTMCampaign = item.CampUTMCampaign
	}

	return &redirect.LinkRedirectTarget{
		Slug:           slug,
		LinkID:         item.ID,
		TenantID:       item.TenantID,
		DestinationURL: item.DestinationURL,
		Status:         "active",
		CampaignID:     item.CampaignID,
		UTMSource:      item.UTMSource,
		UTMMedium:      item.UTMMedium,
		UTMCampaign:    resolvedUTMCampaign,
		UTMTerm:        item.UTMTerm,
		UTMContent:     item.UTMContent,
		CustomDomainID: item.CustomDomainID,
		Hostname:       hostname,
	}, nil
}
