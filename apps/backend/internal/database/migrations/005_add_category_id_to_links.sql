-- Tern Migration 005: Add category_id to links

ALTER TABLE links ADD COLUMN category_id UUID REFERENCES link_categories(id) ON DELETE SET NULL;

---- create above / drop below ----

ALTER TABLE links DROP COLUMN category_id;
