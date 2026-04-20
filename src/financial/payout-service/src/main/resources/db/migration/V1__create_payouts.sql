-- =============================================================
-- V1__create_payouts.sql
-- Initial schema for the payout-service financial domain.
-- =============================================================

CREATE TABLE IF NOT EXISTS payouts (
    id               UUID         NOT NULL DEFAULT gen_random_uuid(),
    vendor_id        UUID         NOT NULL,
    amount           NUMERIC(19, 4) NOT NULL,
    currency         CHAR(3)      NOT NULL DEFAULT 'USD',
    status           VARCHAR(20)  NOT NULL DEFAULT 'PENDING',
    method           VARCHAR(20)  NOT NULL,
    reference        VARCHAR(20)  NOT NULL,
    bank_account     TEXT,
    failure_reason   TEXT,
    scheduled_at     TIMESTAMP,
    processed_at     TIMESTAMP,
    created_at       TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMP    NOT NULL DEFAULT NOW(),

    CONSTRAINT pk_payouts PRIMARY KEY (id),
    CONSTRAINT uq_payout_reference UNIQUE (reference),
    CONSTRAINT chk_payout_status CHECK (
        status IN ('PENDING', 'PROCESSING', 'COMPLETED', 'FAILED', 'CANCELLED')
    ),
    CONSTRAINT chk_payout_method CHECK (
        method IN ('BANK_TRANSFER', 'ACH', 'WIRE', 'PAYPAL', 'CRYPTO')
    ),
    CONSTRAINT chk_amount_positive CHECK (amount > 0)
);

-- Supports paginated vendor payout history
CREATE INDEX idx_payouts_vendor_id     ON payouts (vendor_id);

-- Supports filtering by lifecycle state
CREATE INDEX idx_payouts_status        ON payouts (status);

-- Supports the due-payout batch query (WHERE scheduled_at <= NOW())
CREATE INDEX idx_payouts_scheduled_at  ON payouts (scheduled_at);

-- Composite index speeds up the most common query: vendor + status
CREATE INDEX idx_payouts_vendor_status ON payouts (vendor_id, status);

-- Partial index optimises the batch-processing query for PENDING payouts
CREATE INDEX idx_payouts_pending       ON payouts (scheduled_at)
    WHERE status = 'PENDING';
