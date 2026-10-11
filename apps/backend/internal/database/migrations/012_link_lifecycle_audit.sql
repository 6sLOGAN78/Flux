ALTER TABLE audit_events ADD COLUMN target_link_id uuid;
ALTER TABLE audit_events ADD CONSTRAINT audit_events_scoped_link
    FOREIGN KEY (workspace_id,target_link_id) REFERENCES links(workspace_id,id);

---- create above / drop below ----

ALTER TABLE audit_events DROP COLUMN target_link_id;
