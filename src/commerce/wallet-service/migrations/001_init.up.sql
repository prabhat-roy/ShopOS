-- wallets: one row per customer
CREATE TABLE IF NOT EXISTS wallets (
    id          TEXT           NOT NULL PRIMARY KEY,
    customer_id TEXT           NOT NULL UNIQUE,
    balance     NUMERIC(18, 4) NOT NULL DEFAULT 0 CHECK (balance >= 0),
    currency    TEXT           NOT NULL DEFAULT 'USD',
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_wallets_customer_id ON wallets(customer_id);

-- wallet_transactions: append-only ledger
CREATE TABLE IF NOT EXISTS wallet_transactions (
    id          TEXT           NOT NULL PRIMARY KEY,
    wallet_id   TEXT           NOT NULL REFERENCES wallets(id),
    type        TEXT           NOT NULL CHECK (type IN ('CREDIT', 'DEBIT')),
    amount      NUMERIC(18, 4) NOT NULL CHECK (amount > 0),
    reference   TEXT           NOT NULL DEFAULT '',
    description TEXT           NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_wallet_transactions_wallet_id   ON wallet_transactions(wallet_id);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_created_at  ON wallet_transactions(created_at DESC);
