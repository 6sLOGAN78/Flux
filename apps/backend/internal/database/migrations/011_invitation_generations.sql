-- Resend content has a NEW immutable provider identity; keep historical intent
-- rows instead of changing the identity of an already exposed request.
ALTER TABLE invitations ADD COLUMN version integer NOT NULL DEFAULT 1 CHECK (version > 0);
ALTER TABLE invitations ADD COLUMN issued_at timestamptz NOT NULL DEFAULT now();
UPDATE invitations SET issued_at=created_at;
ALTER TABLE invitation_delivery_intents ADD COLUMN current boolean NOT NULL DEFAULT true;
ALTER TABLE invitation_delivery_intents DROP CONSTRAINT invitation_delivery_intents_workspace_id_invitation_id_key;
CREATE UNIQUE INDEX invitation_delivery_current ON invitation_delivery_intents(workspace_id,invitation_id) WHERE current;

---- create above / drop below ----

-- Refuse to destroy resend history to recreate the old single-intent schema.
DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM invitation_delivery_intents WHERE NOT current) THEN
        RAISE EXCEPTION 'invitation generation history cannot be reversed' USING ERRCODE = '23514';
    END IF;
END $$;
DROP INDEX invitation_delivery_current;
ALTER TABLE invitation_delivery_intents ADD UNIQUE (workspace_id,invitation_id);
ALTER TABLE invitation_delivery_intents DROP COLUMN current;
ALTER TABLE invitations DROP COLUMN issued_at;
ALTER TABLE invitations DROP COLUMN version;
