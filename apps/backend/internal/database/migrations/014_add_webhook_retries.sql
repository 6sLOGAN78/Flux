ALTER TABLE webhook_deliveries ADD COLUMN payload JSONB;
ALTER TABLE webhook_deliveries ADD COLUMN next_attempt_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_retry ON webhook_deliveries(status, next_attempt_at) WHERE status = 'retrying';
