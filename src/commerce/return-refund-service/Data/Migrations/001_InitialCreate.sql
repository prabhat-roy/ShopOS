-- Migration: 001_InitialCreate
-- Service:   return-refund-service
-- Domain:    commerce
-- Created:   2026-04-19

-- Enable pgcrypto for gen_random_uuid() on older Postgres versions.
-- On Postgres 13+ this is built-in; the extension install is idempotent.
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ── return_requests ───────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS return_requests (
    id            UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id      VARCHAR(100)  NOT NULL,
    customer_id   VARCHAR(100)  NOT NULL,
    product_id    VARCHAR(100)  NOT NULL,
    quantity      INT           NOT NULL CHECK (quantity > 0),
    reason        VARCHAR(50)   NOT NULL
                      CHECK (reason IN ('Defective','WrongItem','NotAsDescribed','ChangedMind','Other')),
    notes         VARCHAR(2000) NOT NULL DEFAULT '',
    status        VARCHAR(50)   NOT NULL DEFAULT 'Pending'
                      CHECK (status IN ('Pending','Approved','Rejected','Completed')),
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

-- Indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_return_requests_customer_id
    ON return_requests (customer_id);

CREATE INDEX IF NOT EXISTS idx_return_requests_order_id
    ON return_requests (order_id);

CREATE INDEX IF NOT EXISTS idx_return_requests_status
    ON return_requests (status);

-- ── refund_records ────────────────────────────────────────────────────────
-- One refund per return request (1-to-1).
CREATE TABLE IF NOT EXISTS refund_records (
    id                UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    return_request_id UUID          NOT NULL UNIQUE
                          REFERENCES return_requests (id) ON DELETE CASCADE,
    amount            NUMERIC(12,2) NOT NULL CHECK (amount >= 0),
    currency          CHAR(3)       NOT NULL DEFAULT 'USD',
    method            VARCHAR(50)   NOT NULL DEFAULT 'original'
                          CHECK (method IN ('original','store_credit')),
    processed_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_refund_records_return_request_id
    ON refund_records (return_request_id);

-- ── updated_at trigger ────────────────────────────────────────────────────
-- Automatically refresh updated_at on any UPDATE to return_requests.

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_return_requests_updated_at ON return_requests;
CREATE TRIGGER trg_return_requests_updated_at
    BEFORE UPDATE ON return_requests
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
