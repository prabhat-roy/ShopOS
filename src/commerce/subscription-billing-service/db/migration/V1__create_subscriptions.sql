-- Subscriptions table
CREATE TABLE IF NOT EXISTS subscriptions (
    id               TEXT        PRIMARY KEY,
    customer_id      TEXT        NOT NULL,
    plan_id          TEXT        NOT NULL,
    product_id       TEXT        NOT NULL DEFAULT '',
    status           TEXT        NOT NULL DEFAULT 'active',
    cycle            TEXT        NOT NULL DEFAULT 'monthly',
    price            NUMERIC(12,2) NOT NULL DEFAULT 0,
    currency         TEXT        NOT NULL DEFAULT 'USD',
    trial_ends_at    TIMESTAMPTZ,
    next_billing_at  TIMESTAMPTZ NOT NULL,
    started_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    cancelled_at     TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Billing records table
CREATE TABLE IF NOT EXISTS billing_records (
    id               TEXT        PRIMARY KEY,
    subscription_id  TEXT        NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    amount           NUMERIC(12,2) NOT NULL,
    currency         TEXT        NOT NULL DEFAULT 'USD',
    status           TEXT        NOT NULL DEFAULT 'pending',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_customer_id  ON subscriptions(customer_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status       ON subscriptions(status);
CREATE INDEX IF NOT EXISTS idx_billing_subscription_id    ON billing_records(subscription_id);
