-- bundle-service: initial schema

CREATE TABLE IF NOT EXISTS bundles (
    id          TEXT           PRIMARY KEY,
    name        TEXT           NOT NULL,
    description TEXT           NOT NULL DEFAULT '',
    price       NUMERIC(12, 2) NOT NULL DEFAULT 0,
    currency    TEXT           NOT NULL DEFAULT 'USD',
    items       JSONB          NOT NULL DEFAULT '[]',
    active      BOOLEAN        NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_bundles_active ON bundles(active);
