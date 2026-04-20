-- 001_init.up.sql
-- Creates the wishlist_items table for the wishlist-service.

CREATE TABLE IF NOT EXISTS wishlist_items (
    id           UUID        NOT NULL DEFAULT gen_random_uuid(),
    customer_id  UUID        NOT NULL,
    product_id   TEXT        NOT NULL,
    product_name TEXT        NOT NULL,
    price        NUMERIC(12, 2) NOT NULL DEFAULT 0,
    image_url    TEXT        NOT NULL DEFAULT '',
    added_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT wishlist_items_pkey PRIMARY KEY (customer_id, product_id)
);

-- Index to quickly list items for a customer sorted by recency.
CREATE INDEX IF NOT EXISTS idx_wishlist_items_customer_added
    ON wishlist_items (customer_id, added_at DESC);
