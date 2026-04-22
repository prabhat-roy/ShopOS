-- Flyway migration V002 — Commerce domain schema
-- Services: cart-service, checkout-service, order-service, payment-service,
--           shipping-service, promotions-service, loyalty-service, wallet-service

-- ─── Orders ───────────────────────────────────────────────────────────────────
CREATE TABLE orders (
    id                  UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id             UUID        NOT NULL,
    status              TEXT        NOT NULL DEFAULT 'pending'
                        CHECK (status IN ('pending','confirmed','processing','shipped','delivered','cancelled','refunded')),
    shipping_address    JSONB       NOT NULL,
    billing_address     JSONB,
    subtotal_amount     NUMERIC(12,2) NOT NULL,
    shipping_amount     NUMERIC(12,2) NOT NULL DEFAULT 0,
    tax_amount          NUMERIC(12,2) NOT NULL DEFAULT 0,
    discount_amount     NUMERIC(12,2) NOT NULL DEFAULT 0,
    total_amount        NUMERIC(12,2) NOT NULL,
    currency            CHAR(3)     NOT NULL DEFAULT 'USD',
    coupon_code         TEXT,
    payment_method      TEXT,
    payment_intent_id   TEXT,
    shipping_method     TEXT,
    tracking_number     TEXT,
    notes               TEXT,
    metadata            JSONB       NOT NULL DEFAULT '{}',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_user_id     ON orders (user_id, created_at DESC);
CREATE INDEX idx_orders_status      ON orders (status, created_at DESC);
CREATE INDEX idx_orders_created_at  ON orders (created_at DESC);

CREATE TABLE order_items (
    id                  UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id            UUID        NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id          UUID        NOT NULL,
    variant_id          UUID,
    sku                 TEXT        NOT NULL,
    name                TEXT        NOT NULL,
    quantity            INTEGER     NOT NULL CHECK (quantity > 0),
    unit_price_usd      NUMERIC(12,2) NOT NULL,
    discount_amount_usd NUMERIC(12,2) NOT NULL DEFAULT 0,
    total_price_usd     NUMERIC(12,2) NOT NULL,
    category_id         UUID,
    brand_id            UUID,
    metadata            JSONB       NOT NULL DEFAULT '{}'
);

CREATE INDEX idx_order_items_order_id   ON order_items (order_id);
CREATE INDEX idx_order_items_product_id ON order_items (product_id);

-- ─── Payments ─────────────────────────────────────────────────────────────────
CREATE TABLE payments (
    id                  UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id            UUID        NOT NULL REFERENCES orders(id),
    amount              NUMERIC(12,2) NOT NULL,
    currency            CHAR(3)     NOT NULL DEFAULT 'USD',
    status              TEXT        NOT NULL DEFAULT 'pending'
                        CHECK (status IN ('pending','processing','succeeded','failed','refunded','cancelled')),
    method              TEXT        NOT NULL,
    gateway             TEXT        NOT NULL,
    gateway_transaction_id TEXT     UNIQUE,
    gateway_response    JSONB,
    failure_reason      TEXT,
    refunded_amount     NUMERIC(12,2) DEFAULT 0,
    metadata            JSONB       NOT NULL DEFAULT '{}',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payments_order_id ON payments (order_id);
CREATE INDEX idx_payments_status   ON payments (status, created_at DESC);

-- ─── Promotions ───────────────────────────────────────────────────────────────
CREATE TABLE promotions (
    id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    code            TEXT        UNIQUE,
    name            TEXT        NOT NULL,
    type            TEXT        NOT NULL
                    CHECK (type IN ('percentage','fixed','free_shipping','buy_x_get_y')),
    value           NUMERIC(12,2),
    min_order_amount NUMERIC(12,2),
    max_discount    NUMERIC(12,2),
    usage_limit     INTEGER,
    usage_count     INTEGER     NOT NULL DEFAULT 0,
    per_user_limit  INTEGER     DEFAULT 1,
    applicable_to   TEXT        NOT NULL DEFAULT 'all'
                    CHECK (applicable_to IN ('all','categories','products','users')),
    conditions      JSONB       NOT NULL DEFAULT '{}',
    starts_at       TIMESTAMPTZ,
    ends_at         TIMESTAMPTZ,
    active          BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_promotions_code    ON promotions (code) WHERE active = TRUE;
CREATE INDEX idx_promotions_active  ON promotions (active, ends_at);

-- ─── Loyalty ──────────────────────────────────────────────────────────────────
CREATE TABLE loyalty_accounts (
    id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID        NOT NULL UNIQUE,
    points      INTEGER     NOT NULL DEFAULT 0,
    tier        TEXT        NOT NULL DEFAULT 'bronze'
                CHECK (tier IN ('bronze','silver','gold','platinum')),
    total_earned INTEGER    NOT NULL DEFAULT 0,
    total_redeemed INTEGER  NOT NULL DEFAULT 0,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE loyalty_transactions (
    id          BIGSERIAL   PRIMARY KEY,
    user_id     UUID        NOT NULL,
    points      INTEGER     NOT NULL,
    type        TEXT        NOT NULL CHECK (type IN ('earn','redeem','expire','adjust')),
    reference   TEXT,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_loyalty_tx_user ON loyalty_transactions (user_id, created_at DESC);

-- ─── Wallet ───────────────────────────────────────────────────────────────────
CREATE TABLE wallets (
    id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID        NOT NULL UNIQUE,
    balance     NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (balance >= 0),
    currency    CHAR(3)     NOT NULL DEFAULT 'USD',
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE wallet_transactions (
    id          BIGSERIAL   PRIMARY KEY,
    wallet_id   UUID        NOT NULL REFERENCES wallets(id),
    amount      NUMERIC(12,2) NOT NULL,
    type        TEXT        NOT NULL CHECK (type IN ('credit','debit','refund','payout')),
    reference   TEXT,
    balance_after NUMERIC(12,2) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_wallet_tx ON wallet_transactions (wallet_id, created_at DESC);

-- ─── Triggers ─────────────────────────────────────────────────────────────────
CREATE TRIGGER orders_updated_at   BEFORE UPDATE ON orders   FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER payments_updated_at BEFORE UPDATE ON payments FOR EACH ROW EXECUTE FUNCTION set_updated_at();
