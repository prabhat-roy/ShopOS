CREATE TABLE IF NOT EXISTS shipments (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id         TEXT        NOT NULL,
    customer_id      TEXT        NOT NULL,
    carrier          TEXT        NOT NULL DEFAULT 'fedex',
    tracking_number  TEXT        NOT NULL DEFAULT '',
    status           TEXT        NOT NULL DEFAULT 'pending',
    origin_address   JSONB       NOT NULL DEFAULT '{}',
    dest_address     JSONB       NOT NULL DEFAULT '{}',
    estimated_delivery TIMESTAMPTZ,
    shipped_at       TIMESTAMPTZ,
    delivered_at     TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shipments_order_id ON shipments(order_id);
CREATE INDEX IF NOT EXISTS idx_shipments_status   ON shipments(status);
CREATE INDEX IF NOT EXISTS idx_shipments_customer_id ON shipments(customer_id);
