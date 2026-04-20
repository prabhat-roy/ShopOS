-- inventory-service: initial schema

CREATE TABLE IF NOT EXISTS stock_levels (
    id            TEXT        PRIMARY KEY,
    product_id    TEXT        NOT NULL,
    sku           TEXT        NOT NULL DEFAULT '',
    warehouse_id  TEXT        NOT NULL,
    available     INT         NOT NULL DEFAULT 0,
    reserved      INT         NOT NULL DEFAULT 0,
    reorder_point INT         NOT NULL DEFAULT 10,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (product_id, warehouse_id)
);

CREATE TABLE IF NOT EXISTS reservations (
    id         TEXT        PRIMARY KEY,
    order_id   TEXT        NOT NULL,
    product_id TEXT        NOT NULL,
    quantity   INT         NOT NULL,
    status     TEXT        NOT NULL DEFAULT 'reserved',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_stock_product_id       ON stock_levels(product_id);
CREATE INDEX IF NOT EXISTS idx_reservations_order_id  ON reservations(order_id);
