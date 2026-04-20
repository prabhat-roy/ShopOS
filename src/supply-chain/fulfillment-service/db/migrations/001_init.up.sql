-- 001_init.up.sql — fulfillment-service schema

CREATE TABLE IF NOT EXISTS fulfillments (
    id               TEXT        PRIMARY KEY,
    order_id         TEXT        NOT NULL,
    warehouse_id     TEXT        NOT NULL,
    status           TEXT        NOT NULL
                                 CHECK (status IN ('PENDING','PICKING','PACKING','READY_TO_SHIP','SHIPPED','DELIVERED','CANCELLED'))
                                 DEFAULT 'PENDING',
    tracking_number  TEXT        NOT NULL DEFAULT '',
    carrier          TEXT        NOT NULL DEFAULT '',
    items            JSONB       NOT NULL DEFAULT '[]',
    shipping_address TEXT        NOT NULL DEFAULT '',
    notes            TEXT        NOT NULL DEFAULT '',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_fulfillments_order_id  ON fulfillments (order_id);
CREATE INDEX IF NOT EXISTS idx_fulfillments_status    ON fulfillments (status);
CREATE INDEX IF NOT EXISTS idx_fulfillments_created   ON fulfillments (created_at DESC);
