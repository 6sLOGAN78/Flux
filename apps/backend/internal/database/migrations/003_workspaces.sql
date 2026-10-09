CREATE TABLE workspaces (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL CHECK (char_length(name) BETWEEN 1 AND 100 AND name !~ '^\s*$'),
    created_by uuid NOT NULL REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE memberships (
    workspace_id uuid NOT NULL REFERENCES workspaces(id),
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id),
    role text NOT NULL CHECK (role IN ('owner', 'admin', 'member', 'viewer')),
    created_by uuid NOT NULL REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (workspace_id, id),
    UNIQUE (workspace_id, user_id)
);
CREATE INDEX memberships_user_workspace ON memberships(user_id, workspace_id);

-- Bootstrap has no workspace before creation. Its committed result always
-- resolves to one, and durable user identity scopes the retry key.
CREATE TABLE workspace_bootstrap_requests (
    actor_id uuid NOT NULL REFERENCES users(id),
    request_key text NOT NULL CHECK (length(request_key) BETWEEN 16 AND 128),
    request_hash bytea NOT NULL CHECK (octet_length(request_hash) = 32),
    workspace_id uuid NOT NULL REFERENCES workspaces(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    retain_until timestamptz NOT NULL DEFAULT (now() + interval '24 hours'),
    CHECK (retain_until >= created_at + interval '24 hours'),
    PRIMARY KEY (actor_id, request_key)
);

CREATE TABLE mutation_requests (
    workspace_id uuid NOT NULL REFERENCES workspaces(id),
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    actor_id uuid NOT NULL REFERENCES users(id),
    operation text NOT NULL,
    request_key text NOT NULL CHECK (length(request_key) BETWEEN 16 AND 128),
    request_hash bytea NOT NULL CHECK (octet_length(request_hash) = 32),
    result jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    retain_until timestamptz NOT NULL DEFAULT (now() + interval '24 hours'),
    CHECK (retain_until >= created_at + interval '24 hours'),
    PRIMARY KEY (workspace_id, id),
    UNIQUE (workspace_id, actor_id, operation, request_key)
);

CREATE TABLE audit_events (
    workspace_id uuid NOT NULL REFERENCES workspaces(id),
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    actor_id uuid NOT NULL REFERENCES users(id),
    operation text NOT NULL,
    target_membership_id uuid,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (workspace_id, id)
);
-- Membership targets are historical UUID snapshots, never cascading live FKs.
CREATE FUNCTION preserve_audit_event() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'audit events are immutable' USING ERRCODE = '23514';
END;
$$;
CREATE TRIGGER preserve_audit_event BEFORE UPDATE OR DELETE ON audit_events
    FOR EACH ROW EXECUTE FUNCTION preserve_audit_event();

---- create above / drop below ----

DROP TABLE audit_events;
DROP FUNCTION preserve_audit_event();
DROP TABLE mutation_requests;
DROP TABLE workspace_bootstrap_requests;
DROP TABLE memberships;
DROP TABLE workspaces;
