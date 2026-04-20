-- ============================================================
-- V1__create_accounts.sql
-- Double-entry bookkeeping schema for accounting-service
-- ============================================================

CREATE TABLE IF NOT EXISTS accounts (
    id          UUID            NOT NULL DEFAULT gen_random_uuid(),
    code        VARCHAR(50)     NOT NULL,
    name        VARCHAR(255)    NOT NULL,
    type        VARCHAR(20)     NOT NULL,
    balance     NUMERIC(19, 4)  NOT NULL DEFAULT 0.0000,
    currency    CHAR(3)         NOT NULL DEFAULT 'USD',
    active      BOOLEAN         NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMP       NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP       NOT NULL DEFAULT NOW(),

    CONSTRAINT pk_accounts            PRIMARY KEY (id),
    CONSTRAINT uq_accounts_code       UNIQUE (code),
    CONSTRAINT chk_accounts_type      CHECK (type IN ('ASSET','LIABILITY','EQUITY','REVENUE','EXPENSE')),
    CONSTRAINT chk_accounts_currency  CHECK (currency ~ '^[A-Z]{3}$')
);

CREATE INDEX idx_accounts_type   ON accounts (type);
CREATE INDEX idx_accounts_active ON accounts (active);

-- ─────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS journal_entries (
    id           UUID            NOT NULL DEFAULT gen_random_uuid(),
    reference    VARCHAR(100)    NOT NULL,
    description  VARCHAR(500)    NOT NULL,
    total_amount NUMERIC(19, 4)  NOT NULL,
    currency     CHAR(3)         NOT NULL DEFAULT 'USD',
    created_at   TIMESTAMP       NOT NULL DEFAULT NOW(),

    CONSTRAINT pk_journal_entries          PRIMARY KEY (id),
    CONSTRAINT uq_journal_entries_ref      UNIQUE (reference),
    CONSTRAINT chk_journal_entries_amount  CHECK (total_amount > 0),
    CONSTRAINT chk_journal_entries_curr    CHECK (currency ~ '^[A-Z]{3}$')
);

CREATE INDEX idx_journal_entries_created_at ON journal_entries (created_at);

-- ─────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS journal_lines (
    id         UUID            NOT NULL DEFAULT gen_random_uuid(),
    entry_id   UUID            NOT NULL,
    account_id UUID            NOT NULL,
    type       VARCHAR(10)     NOT NULL,
    amount     NUMERIC(19, 4)  NOT NULL,

    CONSTRAINT pk_journal_lines         PRIMARY KEY (id),
    CONSTRAINT fk_journal_lines_entry   FOREIGN KEY (entry_id)   REFERENCES journal_entries (id) ON DELETE CASCADE,
    CONSTRAINT fk_journal_lines_account FOREIGN KEY (account_id) REFERENCES accounts (id),
    CONSTRAINT chk_journal_lines_type   CHECK (type IN ('debit', 'credit')),
    CONSTRAINT chk_journal_lines_amount CHECK (amount > 0)
);

CREATE INDEX idx_journal_lines_entry_id   ON journal_lines (entry_id);
CREATE INDEX idx_journal_lines_account_id ON journal_lines (account_id);

-- ─────────────────────────────────────────────────────────────
-- Seed chart of accounts (standard COA)
-- ─────────────────────────────────────────────────────────────

INSERT INTO accounts (id, code, name, type, currency) VALUES
    (gen_random_uuid(), 'CASH',        'Cash and Cash Equivalents',  'ASSET',     'USD'),
    (gen_random_uuid(), 'AR',          'Accounts Receivable',        'ASSET',     'USD'),
    (gen_random_uuid(), 'INV',         'Inventory',                  'ASSET',     'USD'),
    (gen_random_uuid(), 'AP',          'Accounts Payable',           'LIABILITY', 'USD'),
    (gen_random_uuid(), 'ACCRUED-EXP', 'Accrued Expenses',           'LIABILITY', 'USD'),
    (gen_random_uuid(), 'RETAINED',    'Retained Earnings',          'EQUITY',    'USD'),
    (gen_random_uuid(), 'COMMON-STK',  'Common Stock',               'EQUITY',    'USD'),
    (gen_random_uuid(), 'SALES-REV',   'Sales Revenue',              'REVENUE',   'USD'),
    (gen_random_uuid(), 'COGS',        'Cost of Goods Sold',         'EXPENSE',   'USD'),
    (gen_random_uuid(), 'OPEX',        'Operating Expenses',         'EXPENSE',   'USD')
ON CONFLICT (code) DO NOTHING;
