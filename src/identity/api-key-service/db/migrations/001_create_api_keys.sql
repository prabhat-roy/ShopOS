-- migrate:up

CREATE TABLE api_keys (
    id          TEXT        PRIMARY KEY,
    owner_id    TEXT        NOT NULL,
    owner_type  TEXT        NOT NULL DEFAULT 'user',
    name        TEXT        NOT NULL,
    key_prefix  TEXT        NOT NULL,
    key_hash    TEXT        NOT NULL UNIQUE,
    scopes      TEXT[]      NOT NULL DEFAULT '{}',
    active      BOOLEAN     NOT NULL DEFAULT TRUE,
    last_used_at TIMESTAMPTZ,
    expires_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_api_keys_owner_id ON api_keys(owner_id);
CREATE INDEX idx_api_keys_key_hash  ON api_keys(key_hash);

-- migrate:down

DROP TABLE IF EXISTS api_keys;
