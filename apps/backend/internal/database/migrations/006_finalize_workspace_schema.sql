-- Tern Migration 006: Finalize Workspace Schema

-- Add owner_id to support the development fallback "Personal Workspaces"
ALTER TABLE workspaces ADD COLUMN owner_id UUID REFERENCES users(id) ON DELETE CASCADE;

-- Migrate existing personal workspaces ownership from workspace_members
UPDATE workspaces w
SET owner_id = wm.user_id
FROM workspace_members wm
WHERE wm.workspace_id = w.id AND w.clerk_org_id IS NULL;

-- Ensure users can only have one personal workspace
CREATE UNIQUE INDEX idx_workspaces_owner_id_personal ON workspaces(owner_id) WHERE clerk_org_id IS NULL;

-- Drop the redundant workspace_members table entirely! Clerk is the source of truth for membership.
DROP TABLE workspace_members;

---- create above / drop below ----

CREATE TABLE workspace_members (
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (workspace_id, user_id)
);

DROP INDEX IF EXISTS idx_workspaces_owner_id_personal;
ALTER TABLE workspaces DROP COLUMN owner_id;
