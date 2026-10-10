CREATE TABLE invitations (
    workspace_id uuid NOT NULL REFERENCES workspaces(id),
    id uuid NOT NULL,
    inviter_id uuid NOT NULL REFERENCES users(id),
    accepted_user_id uuid REFERENCES users(id),
    email text NOT NULL CHECK (length(email) BETWEEN 3 AND 254 AND email = lower(btrim(email))),
    role text NOT NULL CHECK (role IN ('admin','member','viewer')),
    token_digest bytea NOT NULL CHECK (octet_length(token_digest) = 32),
    state text NOT NULL DEFAULT 'pending' CHECK (state IN ('pending','accepted','expired','revoked')),
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL DEFAULT (now() + interval '7 days'),
    CHECK (expires_at > created_at),
    PRIMARY KEY (workspace_id,id)
);
CREATE UNIQUE INDEX invitations_pending_email ON invitations(workspace_id,email) WHERE state='pending';
CREATE INDEX invitations_list ON invitations(workspace_id,id);

-- Durable encrypted intent; API does not claim or deliver it. Lease generation
-- fences future workers, and a successful provider acknowledgement is explicit.
CREATE TABLE invitation_delivery_intents (
    workspace_id uuid NOT NULL,
    id uuid NOT NULL,
    invitation_id uuid NOT NULL,
    key_id text NOT NULL CHECK (key_id ~ '^[A-Za-z0-9_-]{1,64}$'),
    ciphertext bytea NOT NULL CHECK (octet_length(ciphertext) BETWEEN 28 AND 8192),
    state text NOT NULL DEFAULT 'queued' CHECK (state IN ('queued','leased','delivered','failed','cancelled')),
    lease_generation bigint NOT NULL DEFAULT 0 CHECK (lease_generation >= 0),
    lease_until timestamptz,
    attempts integer NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    available_at timestamptz NOT NULL DEFAULT now(),
    delivered_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (workspace_id,id),
    UNIQUE (workspace_id,invitation_id),
    FOREIGN KEY (workspace_id,invitation_id) REFERENCES invitations(workspace_id,id),
    CHECK ((state='leased') = (lease_until IS NOT NULL)),
    CHECK ((state='delivered') = (delivered_at IS NOT NULL))
);
CREATE INDEX invitation_delivery_queue ON invitation_delivery_intents(available_at) WHERE state='queued';
ALTER TABLE audit_events ADD COLUMN target_invitation_id uuid;
ALTER TABLE audit_events ADD FOREIGN KEY (workspace_id,target_invitation_id) REFERENCES invitations(workspace_id,id);

---- create above / drop below ----

ALTER TABLE audit_events DROP COLUMN target_invitation_id;
DROP TABLE invitation_delivery_intents;
DROP TABLE invitations;
