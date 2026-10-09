ALTER TABLE links DROP CONSTRAINT links_short_key_check;
ALTER TABLE links ADD CONSTRAINT links_short_key_check
    CHECK (short_key ~ '^[a-z0-9][a-z0-9_-]{2,63}$' AND
           short_key NOT IN ('api','docs','live','ready','static','login','register','dashboard','settings','links','admin'));

---- create above / drop below ----

-- PostgreSQL validates existing rows; custom keys prevent unsafe rollback.
ALTER TABLE links DROP CONSTRAINT links_short_key_check;
ALTER TABLE links ADD CONSTRAINT links_short_key_check CHECK (short_key ~ '^[a-z2-7]{20}$');
