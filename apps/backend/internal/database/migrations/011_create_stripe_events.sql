-- Tern Migration 011: Stripe events idempotency table

CREATE TABLE IF NOT EXISTS stripe_events (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

---- create above / drop below ----

DROP TABLE IF EXISTS stripe_events;
