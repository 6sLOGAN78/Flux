-- Acknowledged deliveries retain status and provenance without bearer material.
ALTER TABLE invitation_delivery_intents ALTER COLUMN ciphertext DROP NOT NULL;
ALTER TABLE invitation_delivery_intents ADD CONSTRAINT delivery_active_ciphertext
    CHECK (state NOT IN ('queued','leased') OR ciphertext IS NOT NULL);
ALTER TABLE invitation_delivery_intents ADD CONSTRAINT delivery_ack_erased
    CHECK (state <> 'delivered' OR ciphertext IS NULL);

---- create above / drop below ----

-- Erased credentials cannot be reconstructed. Keep this migration irreversible
-- once a delivery has been acknowledged rather than synthesize bearer material.
DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM invitation_delivery_intents WHERE ciphertext IS NULL) THEN
        RAISE EXCEPTION 'delivery erasure cannot be reversed' USING ERRCODE = '23514';
    END IF;
END $$;
ALTER TABLE invitation_delivery_intents DROP CONSTRAINT delivery_ack_erased;
ALTER TABLE invitation_delivery_intents DROP CONSTRAINT delivery_active_ciphertext;
ALTER TABLE invitation_delivery_intents ALTER COLUMN ciphertext SET NOT NULL;
