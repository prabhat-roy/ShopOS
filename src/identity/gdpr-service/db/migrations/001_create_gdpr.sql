-- migrate:up

CREATE TABLE IF NOT EXISTS data_requests (
    id           TEXT        PRIMARY KEY,
    user_id      TEXT        NOT NULL,
    type         TEXT        NOT NULL,
    status       TEXT        NOT NULL DEFAULT 'pending',
    reason       TEXT        NOT NULL DEFAULT '',
    notes        TEXT        NOT NULL DEFAULT '',
    completed_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS consents (
    user_id    TEXT        NOT NULL,
    type       TEXT        NOT NULL,
    granted    BOOLEAN     NOT NULL DEFAULT FALSE,
    ip_address TEXT        NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, type)
);

CREATE INDEX IF NOT EXISTS idx_data_requests_user_id ON data_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_data_requests_status  ON data_requests(status);

-- migrate:down

DROP TABLE IF EXISTS consents;
DROP TABLE IF EXISTS data_requests;
