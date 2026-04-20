-- credit-service: initial schema
-- Up migration: creates credit_accounts and credit_transactions tables.

CREATE TABLE IF NOT EXISTS credit_accounts (
    id               UUID        PRIMARY KEY,
    customer_id      UUID        NOT NULL UNIQUE,
    credit_limit     NUMERIC(18,4) NOT NULL CHECK (credit_limit >= 0),
    available_credit NUMERIC(18,4) NOT NULL CHECK (available_credit >= 0),
    used_credit      NUMERIC(18,4) NOT NULL DEFAULT 0 CHECK (used_credit >= 0),
    currency         VARCHAR(3)  NOT NULL DEFAULT 'USD',
    status           VARCHAR(20) NOT NULL DEFAULT 'active'
                                 CHECK (status IN ('active','suspended','closed')),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_credit_accounts_customer_id
    ON credit_accounts (customer_id);

CREATE INDEX IF NOT EXISTS idx_credit_accounts_status
    ON credit_accounts (status);

CREATE TABLE IF NOT EXISTS credit_transactions (
    id          UUID          PRIMARY KEY,
    account_id  UUID          NOT NULL REFERENCES credit_accounts(id) ON DELETE CASCADE,
    type        VARCHAR(20)   NOT NULL CHECK (type IN ('charge','payment','adjustment')),
    amount      NUMERIC(18,4) NOT NULL,
    reference   VARCHAR(255)  NOT NULL DEFAULT '',
    description TEXT          NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_credit_transactions_account_id
    ON credit_transactions (account_id);

CREATE INDEX IF NOT EXISTS idx_credit_transactions_created_at
    ON credit_transactions (created_at DESC);
