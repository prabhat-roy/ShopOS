CREATE TABLE price_lists (
  id          UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
  name        TEXT    NOT NULL,
  code        TEXT    NOT NULL UNIQUE,
  currency    TEXT    NOT NULL DEFAULT 'USD',
  description TEXT    NOT NULL DEFAULT '',
  active      BOOLEAN NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE price_list_entries (
  id            UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
  price_list_id UUID          NOT NULL REFERENCES price_lists(id) ON DELETE CASCADE,
  product_id    TEXT          NOT NULL,
  price         NUMERIC(12,2) NOT NULL,
  created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
  UNIQUE(price_list_id, product_id)
);

CREATE INDEX idx_price_list_entries_list_id ON price_list_entries(price_list_id);
