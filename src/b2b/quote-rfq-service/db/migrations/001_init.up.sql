-- 001_init.up.sql
-- Creates the quotes table for the quote-rfq-service.

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS quotes (
    id                  UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id              UUID        NOT NULL,
    title               TEXT        NOT NULL,
    description         TEXT        NOT NULL DEFAULT '',
    items               JSONB       NOT NULL DEFAULT '[]',
    requested_delivery  TIMESTAMPTZ NOT NULL,
    status              TEXT        NOT NULL DEFAULT 'DRAFT'
                            CHECK (status IN ('DRAFT','SUBMITTED','UNDER_REVIEW','QUOTED','ACCEPTED','REJECTED','EXPIRED','CANCELLED')),
    total_amount        NUMERIC(18,4) NOT NULL DEFAULT 0,
    currency            CHAR(3)     NOT NULL DEFAULT 'USD',
    valid_until         TIMESTAMPTZ,
    notes               TEXT        NOT NULL DEFAULT '',
    created_by          TEXT        NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_quotes_org_id  ON quotes (org_id);
CREATE INDEX IF NOT EXISTS idx_quotes_status  ON quotes (status);
CREATE INDEX IF NOT EXISTS idx_quotes_created ON quotes (created_at DESC);
