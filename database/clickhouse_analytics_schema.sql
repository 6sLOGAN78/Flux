-- ClickHouse Analytics Schema for Flux Platform
-- Engine: ReplacingMergeTree (deduplicates events by id, ordered by link_id, timestamp)

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
    latitude Nullable(Float64),
    longitude Nullable(Float64),
    user_agent String,
    browser LowCardinality(String),
    browser_version String,
    os LowCardinality(String),
    os_version String,
    device_type LowCardinality(String), -- 'desktop', 'mobile', 'tablet', 'bot'
    referrer_domain String,
    referrer_url String,
    utm_source LowCardinality(String),
    utm_medium LowCardinality(String),
    utm_campaign LowCardinality(String),
    utm_term String,
    utm_content String,
    qr_code_scan UInt8 DEFAULT 0,
    response_time_ms UInt32
) ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(timestamp)
PRIMARY KEY (link_id, timestamp)
ORDER BY (link_id, timestamp, id);
