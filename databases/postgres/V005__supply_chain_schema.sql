-- Supply chain — vendor, warehouse, fulfillment, purchase orders, customs.

CREATE SCHEMA IF NOT EXISTS supply_chain AUTHORIZATION supply_chain_app;
SET search_path TO supply_chain, public;

CREATE TABLE vendor (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code            VARCHAR(32) NOT NULL UNIQUE,
    name            VARCHAR(160) NOT NULL,
    country_iso2    CHAR(2) NOT NULL,
    rating          NUMERIC(3,2) NOT NULL DEFAULT 0 CHECK (rating BETWEEN 0 AND 5),
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE warehouse (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code            VARCHAR(16) NOT NULL UNIQUE,
    name            VARCHAR(120) NOT NULL,
    country_iso2    CHAR(2) NOT NULL,
    region          VARCHAR(64),
    timezone        VARCHAR(64) NOT NULL DEFAULT 'UTC',
    capacity_units  INT,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE purchase_order (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    po_number       VARCHAR(32) NOT NULL UNIQUE,
    vendor_id       UUID NOT NULL REFERENCES vendor(id),
    warehouse_id    UUID NOT NULL REFERENCES warehouse(id),
    status          VARCHAR(24) NOT NULL DEFAULT 'draft',
    total_cents     BIGINT NOT NULL DEFAULT 0,
    currency        CHAR(3) NOT NULL,
    expected_at     DATE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE purchase_order_line (
    po_id           UUID NOT NULL REFERENCES purchase_order(id) ON DELETE CASCADE,
    line_no         INT NOT NULL,
    sku             VARCHAR(64) NOT NULL,
    qty_ordered     INT NOT NULL CHECK (qty_ordered > 0),
    qty_received    INT NOT NULL DEFAULT 0,
    unit_cost_cents BIGINT NOT NULL,
    PRIMARY KEY (po_id, line_no)
);

CREATE TABLE fulfillment (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        UUID NOT NULL,
    warehouse_id    UUID NOT NULL REFERENCES warehouse(id),
    status          VARCHAR(24) NOT NULL DEFAULT 'pending',
    picked_at       TIMESTAMPTZ,
    packed_at       TIMESTAMPTZ,
    shipped_at      TIMESTAMPTZ,
    carrier         VARCHAR(64),
    tracking_no     VARCHAR(64)
);
CREATE INDEX fulfillment_order_idx ON fulfillment (order_id);

CREATE TABLE customs_declaration (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fulfillment_id  UUID NOT NULL REFERENCES fulfillment(id),
    hs_code         VARCHAR(12),
    declared_value_cents BIGINT NOT NULL,
    duty_cents      BIGINT NOT NULL DEFAULT 0,
    origin_country  CHAR(2) NOT NULL,
    destination     CHAR(2) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE supplier_rating_history (
    vendor_id       UUID NOT NULL REFERENCES vendor(id) ON DELETE CASCADE,
    period_start    DATE NOT NULL,
    on_time_pct     NUMERIC(5,2),
    quality_score   NUMERIC(3,2),
    fill_rate_pct   NUMERIC(5,2),
    PRIMARY KEY (vendor_id, period_start)
);

CREATE TABLE route_plan (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id      VARCHAR(64) NOT NULL,
    plan_date       DATE NOT NULL,
    stops           JSONB NOT NULL,
    total_distance_km NUMERIC(8,2),
    total_duration_min INT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
