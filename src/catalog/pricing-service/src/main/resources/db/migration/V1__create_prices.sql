CREATE TABLE prices (
  id          UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
  product_id  TEXT          NOT NULL,
  currency    TEXT          NOT NULL DEFAULT 'USD',
  base_price  NUMERIC(12,2) NOT NULL,
  sale_price  NUMERIC(12,2),
  min_qty     INT           NOT NULL DEFAULT 1,
  segment     TEXT          NOT NULL DEFAULT 'all',
  active      BOOLEAN       NOT NULL DEFAULT TRUE,
  start_at    TIMESTAMPTZ,
  end_at      TIMESTAMPTZ,
  created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_prices_product_id ON prices(product_id, active);
