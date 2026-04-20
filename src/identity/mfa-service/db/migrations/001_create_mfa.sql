-- migrate:up

CREATE TABLE IF NOT EXISTS mfa_setups (
    user_id    TEXT        PRIMARY KEY,
    secret     TEXT        NOT NULL,
    status     TEXT        NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS mfa_backup_codes (
    id        TEXT        PRIMARY KEY,
    user_id   TEXT        NOT NULL REFERENCES mfa_setups(user_id) ON DELETE CASCADE,
    code_hash TEXT        NOT NULL,
    used      BOOLEAN     NOT NULL DEFAULT FALSE,
    used_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_backup_codes_user_id ON mfa_backup_codes(user_id);

-- migrate:down

DROP TABLE IF EXISTS mfa_backup_codes;
DROP TABLE IF EXISTS mfa_setups;
