-- Tern Migration 007: Create campaigns table and add utm/campaign to links

CREATE TABLE IF NOT EXISTS campaigns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    utm_campaign VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_campaigns_workspace FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_campaigns_workspace_id ON campaigns (workspace_id);

DROP TRIGGER IF EXISTS set_updated_at_campaigns ON campaigns;
CREATE TRIGGER set_updated_at_campaigns
    BEFORE UPDATE ON campaigns
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

ALTER TABLE links
    ADD COLUMN IF NOT EXISTS campaign_id UUID REFERENCES campaigns(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS utm_source VARCHAR(255),
    ADD COLUMN IF NOT EXISTS utm_medium VARCHAR(255),
    ADD COLUMN IF NOT EXISTS utm_campaign VARCHAR(255),
    ADD COLUMN IF NOT EXISTS utm_term VARCHAR(255),
    ADD COLUMN IF NOT EXISTS utm_content VARCHAR(255);

CREATE INDEX IF NOT EXISTS idx_links_campaign_id ON links (campaign_id);

---- create above / drop below ----

ALTER TABLE links
    DROP COLUMN IF EXISTS campaign_id,
    DROP COLUMN IF EXISTS utm_source,
    DROP COLUMN IF EXISTS utm_medium,
    DROP COLUMN IF EXISTS utm_campaign,
    DROP COLUMN IF EXISTS utm_term,
    DROP COLUMN IF EXISTS utm_content;

DROP TRIGGER IF EXISTS set_updated_at_campaigns ON campaigns;
DROP TABLE IF EXISTS campaigns;
