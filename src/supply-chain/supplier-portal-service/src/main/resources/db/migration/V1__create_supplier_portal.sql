-- ============================================================
-- V1: Create supplier portal tables
-- ============================================================

-- Supplier Invoices
CREATE TABLE IF NOT EXISTS supplier_invoices (
    id               UUID         NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    vendor_id        UUID         NOT NULL,
    purchase_order_id UUID,
    invoice_number   VARCHAR(100) NOT NULL,
    amount           NUMERIC(19, 4) NOT NULL,
    currency         CHAR(3)      NOT NULL DEFAULT 'USD',
    status           VARCHAR(20)  NOT NULL DEFAULT 'DRAFT',
    due_date         DATE,
    paid_at          TIMESTAMP,
    line_items       TEXT,
    notes            TEXT,
    created_at       TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMP    NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_invoice_number UNIQUE (invoice_number),
    CONSTRAINT chk_invoice_amount CHECK (amount > 0),
    CONSTRAINT chk_invoice_status CHECK (
        status IN ('DRAFT', 'SUBMITTED', 'UNDER_REVIEW', 'APPROVED', 'REJECTED', 'PAID')
    )
);

CREATE INDEX idx_supplier_invoices_vendor_id
    ON supplier_invoices (vendor_id);

CREATE INDEX idx_supplier_invoices_status
    ON supplier_invoices (status);

CREATE INDEX idx_supplier_invoices_vendor_status
    ON supplier_invoices (vendor_id, status);

CREATE INDEX idx_supplier_invoices_created_at
    ON supplier_invoices (created_at DESC);

-- Supplier Catalog Items
CREATE TABLE IF NOT EXISTS supplier_catalog_items (
    id             UUID           NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    vendor_id      UUID           NOT NULL,
    product_id     VARCHAR(200)   NOT NULL,
    sku            VARCHAR(100)   NOT NULL,
    product_name   VARCHAR(500)   NOT NULL,
    unit_price     NUMERIC(19, 4) NOT NULL,
    currency       CHAR(3)        NOT NULL DEFAULT 'USD',
    min_order_qty  INTEGER        NOT NULL DEFAULT 1,
    lead_time_days INTEGER        NOT NULL DEFAULT 0,
    active         BOOLEAN        NOT NULL DEFAULT TRUE,
    updated_at     TIMESTAMP      NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_vendor_sku UNIQUE (vendor_id, sku),
    CONSTRAINT chk_unit_price CHECK (unit_price > 0),
    CONSTRAINT chk_min_order_qty CHECK (min_order_qty >= 1),
    CONSTRAINT chk_lead_time_days CHECK (lead_time_days >= 0)
);

CREATE INDEX idx_supplier_catalog_vendor_id
    ON supplier_catalog_items (vendor_id);

CREATE INDEX idx_supplier_catalog_vendor_active
    ON supplier_catalog_items (vendor_id, active);

CREATE INDEX idx_supplier_catalog_sku
    ON supplier_catalog_items (sku);
