-- V1__create_purchase_orders.sql
-- Creates the purchase_orders and purchase_order_items tables

CREATE TABLE IF NOT EXISTS purchase_orders (
    id                UUID                        NOT NULL DEFAULT gen_random_uuid(),
    vendor_id         UUID                        NOT NULL,
    status            VARCHAR(30)                 NOT NULL DEFAULT 'DRAFT',
    total_amount      NUMERIC(19, 4)              NOT NULL DEFAULT 0,
    currency          CHAR(3)                     NOT NULL DEFAULT 'USD',
    notes             TEXT,
    expected_delivery DATE,
    created_at        TIMESTAMP WITH TIME ZONE    NOT NULL DEFAULT now(),
    updated_at        TIMESTAMP WITH TIME ZONE    NOT NULL DEFAULT now(),

    CONSTRAINT pk_purchase_orders PRIMARY KEY (id),
    CONSTRAINT chk_purchase_orders_status CHECK (
        status IN (
            'DRAFT', 'SUBMITTED', 'APPROVED', 'REJECTED',
            'PARTIALLY_RECEIVED', 'FULLY_RECEIVED', 'CANCELLED'
        )
    ),
    CONSTRAINT chk_purchase_orders_total_amount CHECK (total_amount >= 0)
);

CREATE TABLE IF NOT EXISTS purchase_order_items (
    id           UUID            NOT NULL DEFAULT gen_random_uuid(),
    order_id     UUID            NOT NULL,
    product_id   VARCHAR(255)    NOT NULL,
    sku          VARCHAR(100)    NOT NULL,
    product_name VARCHAR(255)    NOT NULL,
    quantity     INTEGER         NOT NULL,
    unit_price   NUMERIC(19, 4)  NOT NULL,
    total_price  NUMERIC(19, 4)  NOT NULL,
    received_qty INTEGER         NOT NULL DEFAULT 0,

    CONSTRAINT pk_purchase_order_items PRIMARY KEY (id),
    CONSTRAINT fk_poi_order_id
        FOREIGN KEY (order_id) REFERENCES purchase_orders(id) ON DELETE CASCADE,
    CONSTRAINT chk_poi_quantity CHECK (quantity > 0),
    CONSTRAINT chk_poi_unit_price CHECK (unit_price > 0),
    CONSTRAINT chk_poi_total_price CHECK (total_price > 0),
    CONSTRAINT chk_poi_received_qty CHECK (received_qty >= 0)
);

-- Indexes on purchase_orders
CREATE INDEX IF NOT EXISTS idx_po_vendor_id   ON purchase_orders (vendor_id);
CREATE INDEX IF NOT EXISTS idx_po_status      ON purchase_orders (status);
CREATE INDEX IF NOT EXISTS idx_po_created_at  ON purchase_orders (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_po_vendor_status ON purchase_orders (vendor_id, status);

-- Indexes on purchase_order_items
CREATE INDEX IF NOT EXISTS idx_poi_order_id   ON purchase_order_items (order_id);
CREATE INDEX IF NOT EXISTS idx_poi_sku        ON purchase_order_items (sku);
CREATE INDEX IF NOT EXISTS idx_poi_product_id ON purchase_order_items (product_id);

COMMENT ON TABLE purchase_orders IS 'Purchase orders raised against vendors in the supply chain';
COMMENT ON TABLE purchase_order_items IS 'Line items within a purchase order';
COMMENT ON COLUMN purchase_orders.status IS
    'Lifecycle: DRAFT -> SUBMITTED -> APPROVED/REJECTED; APPROVED -> PARTIALLY_RECEIVED -> FULLY_RECEIVED; any cancelable state -> CANCELLED';
COMMENT ON COLUMN purchase_order_items.received_qty IS
    'Running tally of units received; incremented via the receive-items workflow';
