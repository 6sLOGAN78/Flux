CREATE TABLE links (
    workspace_id uuid NOT NULL REFERENCES workspaces(id),
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    creator_user_id uuid NOT NULL REFERENCES users(id),
    managed_host text NOT NULL CHECK (managed_host = lower(managed_host)),
    short_key text NOT NULL CHECK (short_key ~ '^[a-z2-7]{20}$'),
    destination text NOT NULL,
    title text NOT NULL DEFAULT '' CHECK (char_length(title) <= 200),
    lifecycle text NOT NULL DEFAULT 'active' CHECK (lifecycle IN ('active','disabled','archived','deleted')),
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    suspended_at timestamptz,
    suspended_by uuid REFERENCES users(id),
    suspension_reason text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (workspace_id,id),
    CONSTRAINT links_managed_host_short_key_unique UNIQUE (managed_host,short_key),
    CHECK ((suspended_at IS NULL AND suspended_by IS NULL AND suspension_reason IS NULL) OR
           (suspended_at IS NOT NULL AND suspended_by IS NOT NULL AND suspension_reason IS NOT NULL
            AND length(trim(suspension_reason)) > 0))
);

---- create above / drop below ----

DROP TABLE links;
