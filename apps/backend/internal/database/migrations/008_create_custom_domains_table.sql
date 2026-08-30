-- Tern Migration 008: Create custom domains table

CREATE TABLE IF NOT EXISTS custom_domains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    hostname VARCHAR(255) NOT NULL UNIQUE,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    verification_token VARCHAR(255) NOT NULL,
    ssl_status VARCHAR(50) NOT NULL DEFAULT 'none',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_hostname_lowercase CHECK (hostname = LOWER(hostname)),
    CONSTRAINT chk_hostname_no_trailing_dot CHECK (hostname NOT LIKE '%.'),
    CONSTRAINT chk_status_enum CHECK (status IN ('pending', 'verifying', 'active', 'failed', 'disabled')),
    CONSTRAINT chk_ssl_status_enum CHECK (ssl_status IN ('none', 'provisioning', 'active', 'failed'))
);

CREATE INDEX IF NOT EXISTS idx_custom_domains_tenant ON custom_domains(tenant_id);

CREATE TRIGGER set_updated_at_custom_domains
    BEFORE UPDATE ON custom_domains
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

ALTER TABLE links
ADD COLUMN custom_domain_id UUID REFERENCES custom_domains(id) ON DELETE SET NULL;

---- create above / drop below ----

ALTER TABLE links DROP COLUMN custom_domain_id;

DROP TRIGGER IF EXISTS set_updated_at_custom_domains ON custom_domains;
DROP TABLE IF EXISTS custom_domains;
