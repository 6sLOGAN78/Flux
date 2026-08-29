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

func (r *RedisRedirectCache) Get(ctx context.Context, slug string) (*redirect.LinkRedirectTarget, error) {
	key := fmt.Sprintf("link:%s", slug)
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

func (r *RedisRedirectCache) Set(ctx context.Context, slug string, target *redirect.LinkRedirectTarget, ttl time.Duration) error {
	key := fmt.Sprintf("link:%s", slug)
	data, err := json.Marshal(target)
	if err != nil {
		return fmt.Errorf("failed to marshal link for cache: %w", err)
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisRedirectCache) Delete(ctx context.Context, slug string) error {
	key := fmt.Sprintf("link:%s", slug)
	return r.client.Del(ctx, key).Err()
}

type PostgresRedirectRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRedirectRepository(pool *pgxpool.Pool) *PostgresRedirectRepository {
	return &PostgresRedirectRepository{pool: pool}
}

func (r *PostgresRedirectRepository) GetBySlug(ctx context.Context, slug string) (*redirect.LinkRedirectTarget, error) {
	stmt := `
		SELECT id::text, tenant_id::text, destination_url
		FROM links
		WHERE short_code = @slug
	`
	rows, err := r.pool.Query(ctx, stmt, pgx.NamedArgs{"slug": slug})
	if err != nil {
		return nil, sqlerr.HandleError(fmt.Errorf("failed to query redirect: %w", err))
	}
	
	type rowType struct {
		ID             string `db:"id"`
		TenantID       string `db:"tenant_id"`
		DestinationURL string `db:"destination_url"`
	}
	
	item, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[rowType])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, sqlerr.HandleError(fmt.Errorf("failed to collect redirect row: %w", err))
	}

	return &redirect.LinkRedirectTarget{
		Slug:           slug,
		LinkID:         item.ID,
		TenantID:       item.TenantID,
		DestinationURL: item.DestinationURL,
		Status:         "active",
	}, nil
}
