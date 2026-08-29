-- Tern Migration 004: Clerk Auth and Workspaces

CREATE TABLE IF NOT EXISTS workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clerk_org_id VARCHAR(255) UNIQUE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS workspace_members (
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (workspace_id, user_id)
);

-- Safely clear test data as password_hash is being dropped and clerk_user_id is required
TRUNCATE TABLE users CASCADE;

ALTER TABLE users DROP COLUMN IF EXISTS password_hash;
ALTER TABLE users ADD COLUMN clerk_user_id VARCHAR(255) UNIQUE NOT NULL;

-- Add foreign key from links.tenant_id to workspaces.id
-- First, any existing links with non-null tenant_id that don't exist in workspaces will fail.
-- But we checked and there are 0 links. So we can safely add the FK.
ALTER TABLE links
ADD CONSTRAINT fk_links_tenant
FOREIGN KEY (tenant_id)
REFERENCES workspaces(id)
ON DELETE CASCADE;

---- create above / drop below ----

ALTER TABLE links DROP CONSTRAINT IF EXISTS fk_links_tenant;

ALTER TABLE users DROP COLUMN IF EXISTS clerk_user_id;
ALTER TABLE users ADD COLUMN password_hash VARCHAR(255) NOT NULL DEFAULT '';

DROP TABLE IF EXISTS workspace_members;
DROP TABLE IF EXISTS workspaces;
