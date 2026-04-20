CREATE TABLE subscription_plans (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id   TEXT         NOT NULL,
    name         TEXT         NOT NULL,
    description  TEXT         NOT NULL DEFAULT '',
    billing_cycle TEXT        NOT NULL DEFAULT 'MONTHLY',
    price        NUMERIC(12,2) NOT NULL,
    currency     TEXT         NOT NULL DEFAULT 'USD',
    trial_days   INT          NOT NULL DEFAULT 0,
    active       BOOLEAN      NOT NULL DEFAULT TRUE,
    features     TEXT[]       NOT NULL DEFAULT '{}',
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sub_plans_product_id ON subscription_plans(product_id);
