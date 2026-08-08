-- Tern Migration 001: Create links table and updated_at trigger

CREATE OR REPLACE FUNCTION trigger_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    short_code VARCHAR(20) NOT NULL UNIQUE,
    destination_url TEXT NOT NULL,
    tenant_id UUID,
    title TEXT,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_links_short_code ON links (short_code);
CREATE INDEX IF NOT EXISTS idx_links_tenant_id ON links (tenant_id);

DROP TRIGGER IF EXISTS set_updated_at_links ON links;
CREATE TRIGGER set_updated_at_links
    BEFORE UPDATE ON links
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

---- create above / drop below ----

DROP TRIGGER IF EXISTS set_updated_at_links ON links;
DROP TABLE IF EXISTS links;
