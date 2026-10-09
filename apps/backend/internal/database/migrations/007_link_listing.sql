CREATE INDEX links_workspace_created_id_nondeleted
    ON links (workspace_id, created_at DESC, id DESC)
    WHERE lifecycle <> 'deleted';

---- create above / drop below ----

DROP INDEX links_workspace_created_id_nondeleted;
