CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    issuer text NOT NULL CHECK (length(issuer) BETWEEN 1 AND 2048),
    subject text NOT NULL CHECK (length(subject) BETWEEN 6 AND 256),
    verified_email text NOT NULL CHECK (length(verified_email) BETWEEN 3 AND 320),
    profile_verified_at timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (issuer, subject)
);

-- Email is profile data, never an identity key. The namespace and UUID survive
-- profile changes and cannot be reassigned to a different provider identity.
CREATE FUNCTION preserve_user_identity() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.id IS DISTINCT FROM OLD.id OR NEW.issuer IS DISTINCT FROM OLD.issuer
       OR NEW.subject IS DISTINCT FROM OLD.subject THEN
        RAISE EXCEPTION 'user identity is immutable' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER preserve_user_identity BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION preserve_user_identity();

---- create above / drop below ----

DROP TABLE users;
DROP FUNCTION preserve_user_identity();
