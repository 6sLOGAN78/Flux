package db

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/rs/zerolog/log"
)

// InitClickHouse initializes a ClickHouse connection and runs schema migrations.
func InitClickHouse(addr string) (driver.Conn, error) {
	opts, err := clickhouse.ParseDSN(addr)
	if err != nil {
		opts = &clickhouse.Options{
			Addr: []string{addr},
			Auth: clickhouse.Auth{
				Database: "default",
				Username: "default",
				Password: "",
			},
		}
	}
	opts.DialTimeout = 5 * time.Second
	opts.MaxOpenConns = 10
	opts.MaxIdleConns = 5
	opts.ConnMaxLifetime = time.Hour
	
	conn, err := clickhouse.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open clickhouse connection: %w", err)
	}

	var pingErr error
	for i := 0; i < 5; i++ {
		if pingErr = conn.Ping(context.Background()); pingErr == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if pingErr != nil {
		return nil, fmt.Errorf("failed to ping clickhouse after retries: %w", pingErr)
	}

	log.Info().Msg("successfully connected to ClickHouse")
	return conn, nil
}

// MigrateClickHouseSchema creates the analytics_events table if it doesn't exist.
func MigrateClickHouseSchema(ctx context.Context, conn driver.Conn) error {
	// We use standard MergeTree since click events are mostly append-only.
	// Query-time deduplication via argMax or uniq(event_id) is preferred over ReplacingMergeTree to ensure high insert throughput.
	// We partition by toYYYYMM(timestamp) to naturally partition data by month for retention/pruning.
	schema := `
	CREATE TABLE IF NOT EXISTS analytics_events (
		event_id String,
		event_type String,
		timestamp DateTime64(3, 'UTC'),
		link_id String,
		workspace_id String,
		short_code String,
		referrer String,
		user_agent String,
		ip_hash String,
		campaign_id Nullable(String),
		utm_source Nullable(String),
		utm_medium Nullable(String),
		utm_campaign Nullable(String),
		utm_term Nullable(String),
		utm_content Nullable(String),
		custom_domain_id Nullable(String),
		hostname Nullable(String)
	) ENGINE = MergeTree()
	PARTITION BY toYYYYMM(timestamp)
	ORDER BY (workspace_id, link_id, timestamp)
	TTL toDateTime(timestamp) + INTERVAL 90 DAY;
	`
	// TTL is set to 90 days as a standard baseline, satisfying the retention requirement.
	// We order by workspace_id, link_id, timestamp to perfectly optimize for the expected analytics queries.

	if err := conn.Exec(ctx, schema); err != nil {
		return fmt.Errorf("failed to create analytics_events table: %w", err)
	}

	conversionsSchema := `
	CREATE TABLE IF NOT EXISTS conversions (
		conversion_id String,
		workspace_id String,
		timestamp DateTime64(3, 'UTC'),
		conversion_name String,
		revenue Float64,
		currency String,
		click_ids Array(String),
		visitor_id String
	) ENGINE = MergeTree()
	PARTITION BY toYYYYMM(timestamp)
	ORDER BY (workspace_id, timestamp)
	TTL toDateTime(timestamp) + INTERVAL 90 DAY;
	`
	if err := conn.Exec(ctx, conversionsSchema); err != nil {
		return fmt.Errorf("failed to create conversions table: %w", err)
	}

	// For an already existing table, CREATE TABLE IF NOT EXISTS does nothing and ignores new columns.
	// We must explicitly run ALTER TABLE ADD COLUMN IF NOT EXISTS to seamlessly upgrade the schema
	// while keeping existing events backward compatible (new optional/nullable fields).
	alterQueries := []string{
		"ALTER TABLE analytics_events ADD COLUMN IF NOT EXISTS campaign_id Nullable(String)",
		"ALTER TABLE analytics_events ADD COLUMN IF NOT EXISTS utm_source Nullable(String)",
		"ALTER TABLE analytics_events ADD COLUMN IF NOT EXISTS utm_medium Nullable(String)",
		"ALTER TABLE analytics_events ADD COLUMN IF NOT EXISTS utm_campaign Nullable(String)",
		"ALTER TABLE analytics_events ADD COLUMN IF NOT EXISTS utm_term Nullable(String)",
		"ALTER TABLE analytics_events ADD COLUMN IF NOT EXISTS utm_content Nullable(String)",
		"ALTER TABLE analytics_events ADD COLUMN IF NOT EXISTS custom_domain_id Nullable(String)",
		"ALTER TABLE analytics_events ADD COLUMN IF NOT EXISTS hostname Nullable(String)",
		"ALTER TABLE analytics_events ADD INDEX IF NOT EXISTS idx_event_id event_id TYPE bloom_filter(0.01) GRANULARITY 1",
	}

	for _, q := range alterQueries {
		if err := conn.Exec(ctx, q); err != nil {
			return fmt.Errorf("failed to alter analytics_events schema: %w", err)
		}
	}

	log.Info().Msg("successfully applied ClickHouse schema migrations")
	return nil
}
