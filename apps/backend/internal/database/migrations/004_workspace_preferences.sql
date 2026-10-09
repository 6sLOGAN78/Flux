-- Identity-owned preference is a hint, never workspace authority. Membership
-- removal may leave a stale reference; bootstrap must join current membership.
CREATE TABLE workspace_preferences (
    user_id uuid PRIMARY KEY REFERENCES users(id),
    last_workspace_id uuid REFERENCES workspaces(id) ON DELETE SET NULL
);

---- create above / drop below ----

DROP TABLE workspace_preferences;
