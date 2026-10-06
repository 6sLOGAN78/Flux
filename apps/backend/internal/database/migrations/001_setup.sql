-- Bootstrap the migration ledger without introducing product schema.
SELECT 1;

---- create above / drop below ----

-- The bootstrap has no application schema to undo.
SELECT 1;
