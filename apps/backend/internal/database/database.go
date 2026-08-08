// Package database handles PostgreSQL connection pooling and migrations using pgx/v5.
package database

import (
	"context"
	"fmt"
	"time"

	"flux/apps/backend/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

const DatabasePingTimeout = 10 * time.Second

// NewPool initializes a PostgreSQL connection pool using pgx/v5 pgxpool and structured Config.
func NewPool(ctx context.Context, cfg *config.Config, logger *zerolog.Logger) (*pgxpool.Pool, error) {
	if cfg == nil || cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("invalid config or empty DatabaseURL")
	}
	return InitDBPool(ctx, cfg.DatabaseURL)
}

// InitDBPool initializes a PostgreSQL connection pool using pgx/v5 pgxpool.
func InitDBPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database configuration: %w", err)
	}

	// Set connection pool limits
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 15 * time.Minute

	pingCtx, cancel := context.WithTimeout(ctx, DatabasePingTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(pingCtx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgxpool connection pool: %w", err)
	}

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping postgresql database via pgx/v5: %w", err)
	}

	return pool, nil
}
