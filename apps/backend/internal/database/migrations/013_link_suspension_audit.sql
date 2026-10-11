ALTER TABLE audit_events ADD COLUMN reason text
    CHECK (reason IS NULL OR char_length(reason) BETWEEN 1 AND 1000);

---- create above / drop below ----

ALTER TABLE audit_events DROP COLUMN reason;
