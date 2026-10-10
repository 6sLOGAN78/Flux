-- Persist the first possible provider exposure independently of ACK commit.
-- Unknown legacy leased/attempted records use creation as a conservative bound.
ALTER TABLE invitation_delivery_intents ADD COLUMN dispatch_started_at timestamptz;
ALTER TABLE invitation_delivery_intents ADD COLUMN reconciliation_required boolean NOT NULL DEFAULT false;
ALTER TABLE invitation_delivery_intents ADD CONSTRAINT delivery_reconciliation_terminal
    CHECK (NOT reconciliation_required OR state='failed');
UPDATE invitation_delivery_intents SET dispatch_started_at=created_at
    WHERE state='leased' OR attempts>0;

---- create above / drop below ----

DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM invitation_delivery_intents WHERE dispatch_started_at IS NOT NULL) THEN
        RAISE EXCEPTION 'delivery recovery fence cannot be reversed' USING ERRCODE = '23514';
    END IF;
END $$;
ALTER TABLE invitation_delivery_intents DROP CONSTRAINT delivery_reconciliation_terminal;
ALTER TABLE invitation_delivery_intents DROP COLUMN reconciliation_required;
ALTER TABLE invitation_delivery_intents DROP COLUMN dispatch_started_at;
