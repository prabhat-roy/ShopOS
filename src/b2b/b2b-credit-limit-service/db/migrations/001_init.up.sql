-- 001_init.up.sql
-- Creates the org_credit_limits and credit_transactions tables.

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS org_credit_limits (
    id               UUID           PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id           UUID           NOT NULL UNIQUE,
    credit_limit     NUMERIC(18,4)  NOT NULL DEFAULT 0,
    used_credit      NUMERIC(18,4)  NOT NULL DEFAULT 0,
    available_credit NUMERIC(18,4)  NOT NULL DEFAULT 0,
    currency         CHAR(3)        NOT NULL DEFAULT 'USD',
    status           TEXT           NOT NULL DEFAULT 'ACTIVE'
                         CHECK (status IN ('ACTIVE','SUSPENDED','UNDER_REVIEW')),
    risk_score       INT            NOT NULL DEFAULT 50
                         CHECK (risk_score >= 0 AND risk_score <= 100),
    last_reviewed_at TIMESTAMPTZ,
    created_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    -- Invariant: available_credit = credit_limit - used_credit
    CONSTRAINT chk_credit_balance CHECK (available_credit = credit_limit - used_credit),
    CONSTRAINT chk_used_credit     CHECK (used_credit >= 0),
    CONSTRAINT chk_credit_limit    CHECK (credit_limit >= 0)
);

CREATE INDEX IF NOT EXISTS idx_ocl_org_id ON org_credit_limits (org_id);
CREATE INDEX IF NOT EXISTS idx_ocl_status ON org_credit_limits (status);

CREATE TABLE IF NOT EXISTS credit_transactions (
    id         UUID           PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id     UUID           NOT NULL REFERENCES org_credit_limits (org_id) ON DELETE CASCADE,
    type       TEXT           NOT NULL
                   CHECK (type IN ('utilization','payment','adjustment')),
    amount     NUMERIC(18,4)  NOT NULL,
    reference  TEXT           NOT NULL DEFAULT '',
    balance    NUMERIC(18,4)  NOT NULL,
    created_at TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ct_org_id     ON credit_transactions (org_id);
CREATE INDEX IF NOT EXISTS idx_ct_created_at ON credit_transactions (created_at DESC);
