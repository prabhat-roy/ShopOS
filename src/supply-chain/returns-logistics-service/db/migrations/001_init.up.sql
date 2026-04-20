-- Migration: 001_init.up.sql
-- Creates the return_authorizations table for the returns-logistics-service.

BEGIN;

CREATE TABLE IF NOT EXISTS return_authorizations (
    id                  UUID PRIMARY KEY,
    order_id            VARCHAR(255)    NOT NULL,
    customer_id         VARCHAR(255)    NOT NULL,
    -- items stored as JSONB array of {productId, sku, quantity, condition}
    items               JSONB           NOT NULL DEFAULT '[]',
    reason              TEXT            NOT NULL DEFAULT '',
    status              VARCHAR(50)     NOT NULL DEFAULT 'PENDING',
    return_label        TEXT,
    tracking_number     VARCHAR(255),
    warehouse_id        VARCHAR(255),
    inspection_notes    TEXT,
    rejection_reason    TEXT,
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- Indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_return_authorizations_customer_id
    ON return_authorizations (customer_id);

CREATE INDEX IF NOT EXISTS idx_return_authorizations_order_id
    ON return_authorizations (order_id);

CREATE INDEX IF NOT EXISTS idx_return_authorizations_status
    ON return_authorizations (status);

CREATE INDEX IF NOT EXISTS idx_return_authorizations_created_at
    ON return_authorizations (created_at DESC);

-- GIN index enables efficient JSONB queries on items
CREATE INDEX IF NOT EXISTS idx_return_authorizations_items
    ON return_authorizations USING GIN (items);

COMMENT ON TABLE return_authorizations IS
    'Tracks each return authorisation from creation through to final disposition.';

COMMENT ON COLUMN return_authorizations.items IS
    'JSONB array of ReturnItem objects: [{productId, sku, quantity, condition}]';

COMMENT ON COLUMN return_authorizations.status IS
    'Lifecycle status: PENDING | APPROVED | REJECTED | LABEL_ISSUED | IN_TRANSIT | RECEIVED | INSPECTING | COMPLETED | CANCELLED';

COMMIT;
