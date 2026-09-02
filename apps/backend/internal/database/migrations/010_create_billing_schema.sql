-- Tern Migration 010: Billing schema

ALTER TABLE workspaces ADD COLUMN stripe_customer_id VARCHAR(255) UNIQUE;

CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    stripe_subscription_id VARCHAR(255) UNIQUE NOT NULL,
    stripe_customer_id VARCHAR(255) NOT NULL,
    plan_tier VARCHAR(100) NOT NULL DEFAULT 'free',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    current_period_start TIMESTAMPTZ NOT NULL,
    current_period_end TIMESTAMPTZ NOT NULL,
    cancel_at_period_end BOOLEAN NOT NULL DEFAULT false,
    canceled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_subscription_status CHECK (status IN ('active', 'trialing', 'past_due', 'canceled', 'incomplete', 'incomplete_expired', 'unpaid', 'paused'))
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_workspace_id ON subscriptions(workspace_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_stripe_customer_id ON subscriptions(stripe_customer_id);

CREATE TRIGGER set_updated_at_subscriptions
    BEFORE UPDATE ON subscriptions
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

---- create above / drop below ----

DROP TRIGGER IF EXISTS set_updated_at_subscriptions ON subscriptions;
DROP TABLE IF EXISTS subscriptions;
ALTER TABLE workspaces DROP COLUMN IF EXISTS stripe_customer_id;
