CREATE TABLE IF NOT EXISTS webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    endpoint_url TEXT NOT NULL,
    secret TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    events TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_webhooks_workspace_id ON webhooks(workspace_id);
CREATE INDEX IF NOT EXISTS idx_webhooks_worker ON webhooks(workspace_id, active);
