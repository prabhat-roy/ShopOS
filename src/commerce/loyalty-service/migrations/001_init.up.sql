-- loyalty_accounts: one row per customer
CREATE TABLE IF NOT EXISTS loyalty_accounts (
    customer_id  TEXT        NOT NULL PRIMARY KEY,
    points       BIGINT      NOT NULL DEFAULT 0 CHECK (points >= 0),
    tier_name    TEXT        NOT NULL DEFAULT 'Bronze',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- point_transactions: append-only ledger
CREATE TABLE IF NOT EXISTS point_transactions (
    id          TEXT        NOT NULL PRIMARY KEY,
    customer_id TEXT        NOT NULL REFERENCES loyalty_accounts(customer_id),
    type        TEXT        NOT NULL CHECK (type IN ('EARN', 'REDEEM', 'EXPIRE', 'ADJUST')),
    points      BIGINT      NOT NULL,
    balance     BIGINT      NOT NULL,
    order_id    TEXT        NOT NULL DEFAULT '',
    description TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_point_transactions_customer_id ON point_transactions(customer_id);
CREATE INDEX IF NOT EXISTS idx_point_transactions_created_at  ON point_transactions(created_at DESC);
