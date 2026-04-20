-- ─────────────────────────────────────────────────────────────────────────────
-- V1__create_payments.sql
-- ShopOS :: commerce :: payment-service
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS payments (
    id              UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        TEXT            NOT NULL,
    customer_id     TEXT            NOT NULL,
    amount          NUMERIC(12, 2)  NOT NULL,
    currency        TEXT            NOT NULL DEFAULT 'USD',
    status          TEXT            NOT NULL DEFAULT 'PENDING',
    provider        TEXT            NOT NULL DEFAULT 'stripe',
    provider_tx_id  TEXT            NOT NULL DEFAULT '',
    metadata        JSONB           NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_payments_order_id
    ON payments (order_id);

CREATE INDEX IF NOT EXISTS idx_payments_customer_id
    ON payments (customer_id);

CREATE INDEX IF NOT EXISTS idx_payments_status
    ON payments (status);

-- Auto-update updated_at on every row change
CREATE OR REPLACE FUNCTION update_payments_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_payments_updated_at ON payments;

CREATE TRIGGER trg_payments_updated_at
    BEFORE UPDATE ON payments
    FOR EACH ROW
    EXECUTE FUNCTION update_payments_updated_at();
