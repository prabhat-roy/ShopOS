-- Marketplace — sellers, listings, commissions, disputes, payouts.

CREATE SCHEMA IF NOT EXISTS marketplace AUTHORIZATION marketplace_app;
SET search_path TO marketplace, public;

CREATE TABLE seller (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    legal_name      VARCHAR(255) NOT NULL,
    handle          VARCHAR(64) NOT NULL UNIQUE,
    country_iso2    CHAR(2) NOT NULL,
    kyc_status      VARCHAR(24) NOT NULL DEFAULT 'pending',
    payout_method   VARCHAR(32),
    rating          NUMERIC(3,2) NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE listing (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    seller_id       UUID NOT NULL REFERENCES seller(id) ON DELETE CASCADE,
    sku             VARCHAR(64) NOT NULL,
    title           VARCHAR(255) NOT NULL,
    price_cents     BIGINT NOT NULL,
    currency        CHAR(3) NOT NULL,
    qty             INT NOT NULL DEFAULT 0,
    status          VARCHAR(24) NOT NULL DEFAULT 'pending',
    submitted_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    approved_at     TIMESTAMPTZ,
    UNIQUE (seller_id, sku)
);
CREATE INDEX listing_status_idx ON listing (status) WHERE status = 'pending';

CREATE TABLE commission_rule (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_path   LTREE,
    seller_tier     VARCHAR(16),
    rate_pct        NUMERIC(5,2) NOT NULL,
    fixed_cents     BIGINT NOT NULL DEFAULT 0,
    valid_from      DATE NOT NULL,
    valid_to        DATE
);

CREATE TABLE commission_entry (
    id              BIGSERIAL PRIMARY KEY,
    seller_id       UUID NOT NULL REFERENCES seller(id),
    order_id        UUID NOT NULL,
    line_id         UUID NOT NULL,
    gross_cents     BIGINT NOT NULL,
    commission_cents BIGINT NOT NULL,
    currency        CHAR(3) NOT NULL,
    captured_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX commission_seller_idx ON commission_entry (seller_id, captured_at);

CREATE TABLE dispute (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        UUID NOT NULL,
    seller_id       UUID NOT NULL REFERENCES seller(id),
    buyer_id        UUID NOT NULL,
    reason          VARCHAR(64) NOT NULL,
    description     TEXT,
    status          VARCHAR(24) NOT NULL DEFAULT 'open',
    resolution      VARCHAR(64),
    opened_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at       TIMESTAMPTZ
);

CREATE TABLE seller_payout (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    seller_id       UUID NOT NULL REFERENCES seller(id),
    period_start    DATE NOT NULL,
    period_end      DATE NOT NULL,
    gross_cents     BIGINT NOT NULL,
    commission_cents BIGINT NOT NULL,
    fees_cents      BIGINT NOT NULL DEFAULT 0,
    net_cents       BIGINT GENERATED ALWAYS AS (gross_cents - commission_cents - fees_cents) STORED,
    status          VARCHAR(24) NOT NULL DEFAULT 'pending',
    paid_at         TIMESTAMPTZ
);

CREATE TABLE syndication_target (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(64) NOT NULL UNIQUE,
    endpoint        TEXT NOT NULL,
    auth_secret_ref VARCHAR(160),
    is_active       BOOLEAN NOT NULL DEFAULT TRUE
);
