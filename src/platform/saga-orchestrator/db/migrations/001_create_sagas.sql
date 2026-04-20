-- migrate:up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE sagas (
    id           UUID PRIMARY KEY,
    type         TEXT NOT NULL,
    order_id     TEXT NOT NULL,
    state        TEXT NOT NULL,
    steps        JSONB NOT NULL DEFAULT '[]',
    payload      JSONB NOT NULL DEFAULT '{}',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    failed_at    TIMESTAMPTZ,
    error_msg    TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_sagas_order_id ON sagas (order_id);
CREATE INDEX idx_sagas_state    ON sagas (state);

-- migrate:down
DROP TABLE IF EXISTS sagas;
