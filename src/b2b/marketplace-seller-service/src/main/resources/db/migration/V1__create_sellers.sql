-- V1__create_sellers.sql
-- Creates the initial schema for marketplace-seller-service.
-- Managed by Flyway; do not modify after deployment — add new versioned scripts instead.

-- -------------------------------------------------------------------------
-- sellers
-- -------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS sellers (
    id              UUID            NOT NULL DEFAULT gen_random_uuid(),
    org_id          UUID            NOT NULL,
    display_name    VARCHAR(255)    NOT NULL,
    description     TEXT,
    status          VARCHAR(20)     NOT NULL CHECK (status IN ('PENDING','ACTIVE','SUSPENDED','TERMINATED')),
    tier            VARCHAR(20)     NOT NULL DEFAULT 'BRONZE'
                                        CHECK (tier IN ('BRONZE','SILVER','GOLD','PLATINUM')),
    commission_rate NUMERIC(5,2)    NOT NULL DEFAULT 15.00
                                        CHECK (commission_rate >= 0 AND commission_rate <= 100),
    rating          NUMERIC(3,2)    NOT NULL DEFAULT 0.00
                                        CHECK (rating >= 0 AND rating <= 5),
    total_sales     NUMERIC(19,2)   NOT NULL DEFAULT 0.00,
    total_orders    INTEGER         NOT NULL DEFAULT 0 CHECK (total_orders >= 0),
    product_count   INTEGER         NOT NULL DEFAULT 0 CHECK (product_count >= 0),
    return_rate     NUMERIC(5,2)    NOT NULL DEFAULT 0.00
                                        CHECK (return_rate >= 0 AND return_rate <= 100),
    onboarded_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT now(),

    CONSTRAINT pk_sellers           PRIMARY KEY (id),
    CONSTRAINT uq_sellers_org_id    UNIQUE      (org_id)
);

CREATE INDEX idx_sellers_status ON sellers (status);
CREATE INDEX idx_sellers_tier   ON sellers (tier);

-- -------------------------------------------------------------------------
-- seller_products
-- -------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS seller_products (
    id              UUID            NOT NULL DEFAULT gen_random_uuid(),
    seller_id       UUID            NOT NULL,
    product_id      VARCHAR(100)    NOT NULL,
    sku             VARCHAR(100)    NOT NULL,
    seller_sku      VARCHAR(100)    NOT NULL,
    listing_price   NUMERIC(19,2)   NOT NULL CHECK (listing_price > 0),
    status          VARCHAR(20)     NOT NULL DEFAULT 'PENDING'
                                        CHECK (status IN ('ACTIVE','INACTIVE','PENDING')),
    stock_quantity  INTEGER         NOT NULL DEFAULT 0 CHECK (stock_quantity >= 0),
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT now(),

    CONSTRAINT pk_seller_products               PRIMARY KEY (id),
    CONSTRAINT fk_seller_products_seller        FOREIGN KEY (seller_id)
                                                    REFERENCES sellers(id)
                                                    ON DELETE CASCADE,
    CONSTRAINT uq_seller_products_seller_sku    UNIQUE (seller_id, seller_sku)
);

CREATE INDEX idx_sp_seller_id  ON seller_products (seller_id);
CREATE INDEX idx_sp_product_id ON seller_products (product_id);
CREATE INDEX idx_sp_status     ON seller_products (seller_id, status);

-- -------------------------------------------------------------------------
-- Auto-update updated_at via trigger
-- -------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_sellers_updated_at
    BEFORE UPDATE ON sellers
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_seller_products_updated_at
    BEFORE UPDATE ON seller_products
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
