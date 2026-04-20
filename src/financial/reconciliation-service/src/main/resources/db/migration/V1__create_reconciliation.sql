-- ============================================================
-- V1__create_reconciliation.sql
-- Reconciliation records schema
-- ============================================================

CREATE TABLE IF NOT EXISTS reconciliation_records (
    id                      UUID            NOT NULL DEFAULT gen_random_uuid(),
    internal_payment_id     UUID            NOT NULL,
    external_transaction_id VARCHAR(255)    NOT NULL,
    amount                  NUMERIC(19, 4)  NOT NULL,
    currency                CHAR(3)         NOT NULL DEFAULT 'USD',
    internal_amount         NUMERIC(19, 4)  NOT NULL,
    external_amount         NUMERIC(19, 4)  NOT NULL,
    status                  VARCHAR(20)     NOT NULL DEFAULT 'UNMATCHED',
    discrepancy             NUMERIC(19, 4)  NOT NULL DEFAULT 0.0000,
    processor               VARCHAR(100)    NOT NULL,
    reconciled_at           TIMESTAMP,
    notes                   TEXT,
    created_at              TIMESTAMP       NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMP       NOT NULL DEFAULT NOW(),

    CONSTRAINT pk_reconciliation_records         PRIMARY KEY (id),
    CONSTRAINT uq_reconciliation_payment_id      UNIQUE (internal_payment_id),
    CONSTRAINT chk_reconciliation_status         CHECK (status IN ('MATCHED','UNMATCHED','DISPUTED','RESOLVED')),
    CONSTRAINT chk_reconciliation_currency       CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT chk_reconciliation_amount         CHECK (amount > 0),
    CONSTRAINT chk_reconciliation_discrepancy    CHECK (discrepancy >= 0)
);

CREATE INDEX idx_recon_status          ON reconciliation_records (status);
CREATE INDEX idx_recon_processor       ON reconciliation_records (processor);
CREATE INDEX idx_recon_created_at      ON reconciliation_records (created_at);
CREATE INDEX idx_recon_processor_date  ON reconciliation_records (processor, created_at);
CREATE INDEX idx_recon_ext_tx_id       ON reconciliation_records (external_transaction_id);
