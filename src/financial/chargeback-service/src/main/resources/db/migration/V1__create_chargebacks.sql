-- V1__create_chargebacks.sql
-- Chargeback dispute lifecycle table

CREATE TABLE IF NOT EXISTS chargebacks (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id            VARCHAR(255) NOT NULL,
    order_id              VARCHAR(255) NOT NULL,
    customer_id           VARCHAR(255) NOT NULL,
    amount                NUMERIC(19, 4) NOT NULL,
    currency              CHAR(3) NOT NULL,
    status                VARCHAR(30) NOT NULL DEFAULT 'OPEN',
    reason_code           VARCHAR(50),
    reason_description    TEXT,
    evidence_due_date     TIMESTAMPTZ,
    evidence_submitted_at TIMESTAMPTZ,
    resolved_at           TIMESTAMPTZ,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chargebacks_payment_id   ON chargebacks (payment_id);
CREATE INDEX IF NOT EXISTS idx_chargebacks_customer_id  ON chargebacks (customer_id);
CREATE INDEX IF NOT EXISTS idx_chargebacks_status       ON chargebacks (status);
CREATE INDEX IF NOT EXISTS idx_chargebacks_created_at   ON chargebacks (created_at DESC);
