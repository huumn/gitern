CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

DO $$ BEGIN
  CREATE DOMAIN email AS citext
    CHECK ( value ~ '^[a-zA-Z0-9.!#$%&''*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$' );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
  CREATE DOMAIN accname AS citext
    CHECK ( value ~ '^(?=.{1,40}$)(?![-_])(?!.*[-_]{2})[a-zA-Z0-9-]+(?<![-_])$' );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE keytype AS ENUM ('ssh-rsa', 'ssh-ed25519', 'ecdsa-sha2-nistp256', 'ecdsa-sha2-nistp384', 'ecdsa-sha2-nistp521');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE DOMAIN base64 AS text
      CHECK (value ~  '^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE DOMAIN sessionID AS text
      CHECK (value ~  '^[A-Za-z0-9_-]{4,}$');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE DOMAIN fingerprint AS text
      CHECK (value ~  '^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2,3})?$');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE stripestatus AS ENUM ('paid', 'failed');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE OR REPLACE FUNCTION pubkey2fingerprint(base64) RETURNS fingerprint
    AS $$ SELECT cast(rtrim(encode(digest(decode($1, 'base64'), 'sha256'), 'base64'), '=') AS fingerprint); $$
    LANGUAGE SQL
    IMMUTABLE
    RETURNS NULL ON NULL INPUT;

CREATE TABLE IF NOT EXISTS accounts
(
    name accname PRIMARY KEY,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    stripe_id text,
    stripe_status stripestatus
);

ALTER TABLE accounts DROP COLUMN IF EXISTS quota;
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS stripe_id TEXT;
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS stripe_status stripestatus;

CREATE TABLE IF NOT EXISTS pubkeys
(
    fingerprint fingerprint PRIMARY KEY CHECK (char_length(fingerprint) <= 44),
    keytype keytype NOT NULL,
    pubkey base64 NOT NULL,
    comment text CHECK (comment !~ ' '),
    created_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS accounts_pubkeys 
(
    account_name accname REFERENCES accounts(name),
    pubkey_fingerprint fingerprint REFERENCES pubkeys(fingerprint),
    PRIMARY KEY (account_name, pubkey_fingerprint)
);

CREATE TABLE IF NOT EXISTS sessions 
(
    id sessionID NOT NULL,
    name accname NOT NULL,
    fingerprint fingerprint NOT NULL,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    FOREIGN KEY (name, fingerprint) REFERENCES accounts_pubkeys (account_name, pubkey_fingerprint) ON DELETE CASCADE,
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX IF NOT EXISTS sessions_name_fingerprint_idx ON sessions (name, fingerprint);

CREATE OR REPLACE FUNCTION gen_session_id(integer) RETURNS sessionID
    AS $$ SELECT cast(translate(rtrim(encode(gen_random_bytes($1), 'base64'), '='), '+/', '-_') AS sessionID); $$
    LANGUAGE SQL
    IMMUTABLE
    RETURNS NULL ON NULL INPUT;

CREATE OR REPLACE FUNCTION delete_sessions(accname, fingerprint) RETURNS void
    AS $$ DELETE FROM sessions WHERE (name = $1 AND fingerprint = $2) OR NOW() - created_at > interval '5 minutes'; $$
    LANGUAGE SQL
    VOLATILE
    RETURNS NULL ON NULL INPUT;

CREATE OR REPLACE FUNCTION create_session(accname, fingerprint) RETURNS sessionID
AS $$
DECLARE
    newID sessionID;
BEGIN
    PERFORM delete_sessions($1, $2);
    LOOP
        newID := gen_session_id(6);
        BEGIN
            INSERT INTO sessions (id, name, fingerprint) VALUES (newID, $1, $2);
            EXIT;
        EXCEPTION WHEN unique_violation THEN

        END;
    END LOOP;
    RETURN newID;
END;
$$ LANGUAGE PLPGSQL;