-- Tern Migration 002: Create link categories and attachments

CREATE TABLE IF NOT EXISTS link_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID,
    name TEXT NOT NULL,
    color TEXT DEFAULT '#3b82f6',
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_link_categories_tenant_id ON link_categories(tenant_id);
CREATE UNIQUE INDEX IF NOT EXISTS link_categories_unique_name ON link_categories(tenant_id, name);

DROP TRIGGER IF EXISTS set_updated_at_link_categories ON link_categories;
CREATE TRIGGER set_updated_at_link_categories
    BEFORE UPDATE ON link_categories
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TABLE IF NOT EXISTS link_attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    link_id UUID NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    uploaded_by TEXT NOT NULL,
    download_key TEXT NOT NULL,
    file_size BIGINT,
    mime_type TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_todo_attachments_link_id ON link_attachments(link_id);
CREATE INDEX IF NOT EXISTS idx_todo_attachments_uploaded_by ON link_attachments(uploaded_by);

DROP TRIGGER IF EXISTS set_updated_at_link_attachments ON link_attachments;
CREATE TRIGGER set_updated_at_link_attachments
    BEFORE UPDATE ON link_attachments
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

---- create above / drop below ----

DROP TRIGGER IF EXISTS set_updated_at_link_attachments ON link_attachments;
DROP TABLE IF EXISTS link_attachments;

DROP TRIGGER IF EXISTS set_updated_at_link_categories ON link_categories;
DROP TABLE IF EXISTS link_categories;
