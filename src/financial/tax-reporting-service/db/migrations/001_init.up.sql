-- Migration: 001_init.up.sql
-- Creates the tax_records table for the tax-reporting-service.

CREATE TABLE IF NOT EXISTS tax_records (
    id               TEXT        NOT NULL PRIMARY KEY,
    order_id         TEXT        NOT NULL,
    customer_id      TEXT        NOT NULL,
    jurisdiction     TEXT        NOT NULL,
    tax_type         TEXT        NOT NULL CHECK (tax_type IN ('VAT', 'GST', 'SALES_TAX', 'EXCISE')),
    taxable_amount   NUMERIC(18, 4) NOT NULL CHECK (taxable_amount >= 0),
    tax_rate         NUMERIC(7, 4)  NOT NULL CHECK (tax_rate >= 0 AND tax_rate <= 100),
    tax_amount       NUMERIC(18, 4) NOT NULL CHECK (tax_amount >= 0),
    currency         CHAR(3)     NOT NULL,
    transaction_date TIMESTAMPTZ NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for filtering and aggregation by jurisdiction.
CREATE INDEX IF NOT EXISTS idx_tax_records_jurisdiction
    ON tax_records (jurisdiction);

-- Index for filtering by tax type.
CREATE INDEX IF NOT EXISTS idx_tax_records_tax_type
    ON tax_records (tax_type);

-- Index for date-range queries and monthly aggregation.
CREATE INDEX IF NOT EXISTS idx_tax_records_transaction_date
    ON tax_records (transaction_date);

-- Composite index for summary queries (jurisdiction + date).
CREATE INDEX IF NOT EXISTS idx_tax_records_jurisdiction_date
    ON tax_records (jurisdiction, transaction_date);

-- Composite index for tax type + date filtering.
CREATE INDEX IF NOT EXISTS idx_tax_records_tax_type_date
    ON tax_records (tax_type, transaction_date);
