CREATE TABLE orders (
  id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  customer_id      TEXT         NOT NULL,
  status           TEXT         NOT NULL DEFAULT 'PENDING',
  subtotal         NUMERIC(12,2) NOT NULL DEFAULT 0,
  tax              NUMERIC(12,2) NOT NULL DEFAULT 0,
  shipping         NUMERIC(12,2) NOT NULL DEFAULT 0,
  total            NUMERIC(12,2) NOT NULL DEFAULT 0,
  currency         TEXT         NOT NULL DEFAULT 'USD',
  shipping_address JSONB        NOT NULL DEFAULT '{}',
  notes            TEXT         NOT NULL DEFAULT '',
  created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE order_items (
  id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id   UUID         NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  product_id TEXT         NOT NULL,
  sku        TEXT         NOT NULL DEFAULT '',
  name       TEXT         NOT NULL,
  price      NUMERIC(12,2) NOT NULL,
  quantity   INT          NOT NULL,
  total      NUMERIC(12,2) NOT NULL
);

CREATE INDEX idx_orders_customer_id   ON orders(customer_id);
CREATE INDEX idx_orders_status        ON orders(status);
CREATE INDEX idx_order_items_order_id ON order_items(order_id);
