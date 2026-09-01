-- +goose Up
ALTER TABLE workspaces ADD COLUMN tracking_client_id UUID DEFAULT gen_random_uuid();
ALTER TABLE workspaces ADD CONSTRAINT unique_tracking_client_id UNIQUE (tracking_client_id);

-- +goose Down
ALTER TABLE workspaces DROP COLUMN tracking_client_id;
