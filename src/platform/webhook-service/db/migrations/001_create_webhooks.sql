-- migrate:up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE webhooks (
    id          TEXT        PRIMARY KEY,
    owner_id    TEXT        NOT NULL,
    url         TEXT        NOT NULL,
    events      TEXT[]      NOT NULL DEFAULT '{}',
    secret      TEXT        NOT NULL DEFAULT '',
    active      BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhooks_owner_id ON webhooks (owner_id);
CREATE INDEX idx_webhooks_events   ON webhooks USING GIN (events);

CREATE TABLE webhook_deliveries (
    id          TEXT        PRIMARY KEY,
    webhook_id  TEXT        NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    event_topic TEXT        NOT NULL,
    payload     JSONB       NOT NULL DEFAULT '{}',
    status_code INT         NOT NULL DEFAULT 0,
    attempt     INT         NOT NULL DEFAULT 1,
    success     BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_deliveries_webhook_id ON webhook_deliveries (webhook_id);
CREATE INDEX idx_deliveries_created_at ON webhook_deliveries (created_at);

-- migrate:down
DROP TABLE IF EXISTS webhook_deliveries;
DROP TABLE IF EXISTS webhooks;
