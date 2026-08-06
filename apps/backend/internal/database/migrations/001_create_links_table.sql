-- Tern Migration 001: Create links table
-- Write your migrate up statement here

CREATE TABLE IF NOT EXISTS links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    short_code VARCHAR(20) NOT NULL UNIQUE,
    destination_url TEXT NOT NULL,
    tenant_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_links_short_code ON links (short_code);

---- create above / drop below ----

DROP TABLE IF EXISTS links;
