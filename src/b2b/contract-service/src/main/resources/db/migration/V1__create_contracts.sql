-- ─────────────────────────────────────────────────────────────────────────────
-- V1__create_contracts.sql
-- Creates the contracts table for the contract-service
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS contracts (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id               UUID          NOT NULL,
    vendor_id            UUID,
    title                VARCHAR(500)  NOT NULL,
    type                 VARCHAR(50)   NOT NULL,
    status               VARCHAR(50)   NOT NULL DEFAULT 'DRAFT',
    description          TEXT,
    terms                TEXT,
    value                NUMERIC(19, 4),
    currency             CHAR(3)       NOT NULL DEFAULT 'USD',
    start_date           DATE          NOT NULL,
    end_date             DATE          NOT NULL,
    auto_renew           BOOLEAN       NOT NULL DEFAULT FALSE,
    signed_by_buyer      BOOLEAN       NOT NULL DEFAULT FALSE,
    signed_by_vendor     BOOLEAN       NOT NULL DEFAULT FALSE,
    signed_at            TIMESTAMPTZ,
    termination_reason   TEXT,
    created_by           VARCHAR(255)  NOT NULL,
    created_at           TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ   NOT NULL DEFAULT now(),

    CONSTRAINT chk_contracts_dates CHECK (end_date >= start_date)
);

CREATE INDEX IF NOT EXISTS idx_contracts_org_id     ON contracts (org_id);
CREATE INDEX IF NOT EXISTS idx_contracts_status     ON contracts (status);
CREATE INDEX IF NOT EXISTS idx_contracts_org_status ON contracts (org_id, status);
CREATE INDEX IF NOT EXISTS idx_contracts_end_date   ON contracts (end_date, status);
CREATE INDEX IF NOT EXISTS idx_contracts_created_at ON contracts (created_at DESC);
