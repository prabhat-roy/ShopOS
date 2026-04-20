-- =============================================================
-- V1__create_invoices.sql
-- Initial schema for the invoice-service financial domain.
-- =============================================================

CREATE TABLE IF NOT EXISTS invoices (
    id               UUID         NOT NULL DEFAULT gen_random_uuid(),
    order_id         UUID         NOT NULL,
    customer_id      UUID         NOT NULL,
    invoice_number   VARCHAR(20)  NOT NULL,
    status           VARCHAR(20)  NOT NULL DEFAULT 'DRAFT',
    subtotal         NUMERIC(19, 4) NOT NULL,
    tax_amount       NUMERIC(19, 4) NOT NULL,
    total_amount     NUMERIC(19, 4) NOT NULL,
    currency         CHAR(3)      NOT NULL DEFAULT 'USD',
    line_items       TEXT,
    billing_address  TEXT,
    due_date         DATE         NOT NULL,
    paid_at          TIMESTAMP,
    notes            TEXT,
    created_at       TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMP    NOT NULL DEFAULT NOW(),

    CONSTRAINT pk_invoices PRIMARY KEY (id),
    CONSTRAINT uq_invoice_number UNIQUE (invoice_number),
    CONSTRAINT chk_invoice_status CHECK (
        status IN ('DRAFT', 'ISSUED', 'SENT', 'PAID', 'OVERDUE', 'CANCELLED', 'VOID')
    ),
    CONSTRAINT chk_subtotal_positive    CHECK (subtotal    >= 0),
    CONSTRAINT chk_tax_amount_positive  CHECK (tax_amount  >= 0),
    CONSTRAINT chk_total_amount_positive CHECK (total_amount > 0)
);

-- Supports fast lookup by the upstream order reference
CREATE INDEX idx_invoices_order_id     ON invoices (order_id);

-- Supports paginated customer invoice history
CREATE INDEX idx_invoices_customer_id  ON invoices (customer_id);

-- Supports filtering / querying by lifecycle state
CREATE INDEX idx_invoices_status       ON invoices (status);

-- Supports the overdue-detection batch query (WHERE due_date < today)
CREATE INDEX idx_invoices_due_date     ON invoices (due_date);

-- Composite index speeds up the most common query: customer + status
CREATE INDEX idx_invoices_customer_status ON invoices (customer_id, status);
