-- Catalog domain — relational data for catalog services that depend on Postgres
-- Owners: catalog team. Services: category-service, brand-service, pricing-service,
-- inventory-service, bundle-service, subscription-product-service, price-list-service,
-- product-label-service, variant-service, product-import-service.

CREATE SCHEMA IF NOT EXISTS catalog AUTHORIZATION catalog_app;
SET search_path TO catalog, public;

CREATE TABLE category (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id       UUID REFERENCES category(id) ON DELETE SET NULL,
    slug            VARCHAR(160) NOT NULL UNIQUE,
    name            VARCHAR(160) NOT NULL,
    path            LTREE NOT NULL,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order      INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX category_path_gist ON category USING GIST (path);
CREATE INDEX category_parent_idx ON category (parent_id);

CREATE TABLE brand (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug            VARCHAR(160) NOT NULL UNIQUE,
    name            VARCHAR(160) NOT NULL,
    logo_url        TEXT,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE price (
    sku             VARCHAR(64) NOT NULL,
    currency        CHAR(3) NOT NULL,
    price_cents     BIGINT NOT NULL CHECK (price_cents >= 0),
    list_cents      BIGINT,
    valid_from      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    valid_to        TIMESTAMPTZ,
    region          VARCHAR(8) NOT NULL DEFAULT 'GLOBAL',
    PRIMARY KEY (sku, currency, region, valid_from)
);
CREATE INDEX price_sku_currency_idx ON price (sku, currency, region) WHERE valid_to IS NULL;

CREATE TABLE price_list (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(160) NOT NULL,
    customer_group  VARCHAR(64),
    currency        CHAR(3) NOT NULL,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE inventory (
    sku             VARCHAR(64) NOT NULL,
    warehouse_id    UUID NOT NULL,
    on_hand         INT NOT NULL DEFAULT 0,
    reserved        INT NOT NULL DEFAULT 0,
    available       INT GENERATED ALWAYS AS (on_hand - reserved) STORED,
    reorder_point   INT NOT NULL DEFAULT 0,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (sku, warehouse_id)
);
CREATE INDEX inventory_low_stock_idx ON inventory (sku) WHERE available < reorder_point;

CREATE TABLE bundle (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sku             VARCHAR(64) NOT NULL UNIQUE,
    name            VARCHAR(160) NOT NULL,
    discount_pct    NUMERIC(5,2) NOT NULL DEFAULT 0,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE bundle_item (
    bundle_id       UUID NOT NULL REFERENCES bundle(id) ON DELETE CASCADE,
    sku             VARCHAR(64) NOT NULL,
    qty             INT NOT NULL CHECK (qty > 0),
    PRIMARY KEY (bundle_id, sku)
);

CREATE TABLE subscription_product (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sku             VARCHAR(64) NOT NULL UNIQUE,
    interval        VARCHAR(16) NOT NULL CHECK (interval IN ('weekly','monthly','quarterly','yearly')),
    price_cents     BIGINT NOT NULL,
    currency        CHAR(3) NOT NULL,
    trial_days      INT NOT NULL DEFAULT 0,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE product_label (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sku             VARCHAR(64) NOT NULL,
    label           VARCHAR(64) NOT NULL,
    color_hex       CHAR(7),
    valid_from      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    valid_to        TIMESTAMPTZ
);
CREATE INDEX product_label_sku_idx ON product_label (sku);

CREATE TABLE variant (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_sku      VARCHAR(64) NOT NULL,
    sku             VARCHAR(64) NOT NULL UNIQUE,
    attributes      JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX variant_parent_idx ON variant (parent_sku);

CREATE TABLE import_job (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source          VARCHAR(64) NOT NULL,
    status          VARCHAR(24) NOT NULL DEFAULT 'pending',
    total_rows      INT,
    processed_rows  INT NOT NULL DEFAULT 0,
    error_count     INT NOT NULL DEFAULT 0,
    started_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at     TIMESTAMPTZ
);
