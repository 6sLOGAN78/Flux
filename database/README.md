# Database Architecture & Schemas (`database/`)

Flux uses a **Polyglot Persistence Layer**:
1. **PostgreSQL**: Transactional relational store for links, users, workspaces, domains, campaigns, and webhooks.
2. **ClickHouse**: High-throughput time-series database for append-only click event analytics.

---

## 🐘 1. PostgreSQL Schema & Migrations

PostgreSQL versioned migration scripts are stored in [`apps/backend/internal/database/migrations/`](file:///home/logan78/Desktop/flux/apps/backend/internal/database/migrations/):

- `001_create_links_table.sql`: Creates `links` table with UUID primary keys, `short_code UNIQUE` constraint, indexes on `short_code` & `tenant_id`, and `updated_at` triggers.
- `002_create_link_categories_and_attachments.sql`: Creates link categories, tags, and category attachment tables.

---

## ⚡ 2. ClickHouse Analytics Schema

Defined in [`database/clickhouse_analytics_schema.sql`](file:///home/logan78/Desktop/flux/database/clickhouse_analytics_schema.sql):

```sql
CREATE TABLE IF NOT EXISTS click_events (
    id UUID,
    link_id UUID,
    domain_id Nullable(UUID),
    user_id UUID,
    timestamp DateTime64(3, 'UTC'),
    ip_address String,
    country_code LowCardinality(String),
    region String,
    city String,
    user_agent String,
    browser LowCardinality(String),
    os LowCardinality(String),
    device_type LowCardinality(String),
    referrer_domain String,
    utm_source LowCardinality(String),
    utm_medium LowCardinality(String),
    utm_campaign LowCardinality(String),
    qr_code_scan UInt8 DEFAULT 0,
    response_time_ms UInt32
) ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(timestamp)
PRIMARY KEY (link_id, timestamp)
ORDER BY (link_id, timestamp, id);
```
