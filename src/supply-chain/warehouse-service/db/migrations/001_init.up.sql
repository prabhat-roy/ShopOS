-- 001_init.up.sql — warehouse-service schema

CREATE TABLE IF NOT EXISTS warehouses (
    id          TEXT        PRIMARY KEY,
    name        TEXT        NOT NULL,
    location    TEXT        NOT NULL DEFAULT '',
    address     TEXT        NOT NULL DEFAULT '',
    capacity    INTEGER     NOT NULL DEFAULT 0 CHECK (capacity >= 0),
    active      BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_warehouses_active ON warehouses (active);

CREATE TABLE IF NOT EXISTS stock_movements (
    id            TEXT        PRIMARY KEY,
    warehouse_id  TEXT        NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
    product_id    TEXT        NOT NULL,
    sku           TEXT        NOT NULL,
    movement_type TEXT        NOT NULL CHECK (movement_type IN ('inbound', 'outbound')),
    quantity      INTEGER     NOT NULL CHECK (quantity > 0),
    reference_id  TEXT        NOT NULL DEFAULT '',
    notes         TEXT        NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_stock_movements_warehouse  ON stock_movements (warehouse_id);
CREATE INDEX IF NOT EXISTS idx_stock_movements_product    ON stock_movements (warehouse_id, product_id);
CREATE INDEX IF NOT EXISTS idx_stock_movements_created_at ON stock_movements (created_at DESC);
