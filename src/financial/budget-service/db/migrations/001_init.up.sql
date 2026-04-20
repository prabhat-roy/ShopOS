-- budget-service: initial schema
-- Up migration: creates budgets, budget_allocations and spending_records tables.

CREATE TABLE IF NOT EXISTS budgets (
    id               UUID          PRIMARY KEY,
    department       VARCHAR(255)  NOT NULL,
    name             VARCHAR(255)  NOT NULL,
    period           VARCHAR(20)   NOT NULL CHECK (period IN ('MONTHLY','QUARTERLY','ANNUAL')),
    fiscal_year      INT           NOT NULL,
    start_date       TIMESTAMPTZ   NOT NULL,
    end_date         TIMESTAMPTZ   NOT NULL,
    total_amount     NUMERIC(18,4) NOT NULL CHECK (total_amount >= 0),
    allocated_amount NUMERIC(18,4) NOT NULL DEFAULT 0 CHECK (allocated_amount >= 0),
    spent_amount     NUMERIC(18,4) NOT NULL DEFAULT 0 CHECK (spent_amount >= 0),
    remaining_amount NUMERIC(18,4) NOT NULL CHECK (remaining_amount >= 0),
    currency         VARCHAR(3)    NOT NULL DEFAULT 'USD',
    status           VARCHAR(20)   NOT NULL DEFAULT 'DRAFT'
                                   CHECK (status IN ('DRAFT','ACTIVE','CLOSED')),
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_budgets_department  ON budgets (department);
CREATE INDEX IF NOT EXISTS idx_budgets_status       ON budgets (status);
CREATE INDEX IF NOT EXISTS idx_budgets_fiscal_year  ON budgets (fiscal_year);

CREATE TABLE IF NOT EXISTS budget_allocations (
    id               UUID          PRIMARY KEY,
    budget_id        UUID          NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
    category         VARCHAR(255)  NOT NULL,
    allocated_amount NUMERIC(18,4) NOT NULL CHECK (allocated_amount > 0),
    spent_amount     NUMERIC(18,4) NOT NULL DEFAULT 0 CHECK (spent_amount >= 0),
    notes            TEXT          NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_budget_allocations_budget_id
    ON budget_allocations (budget_id);

CREATE TABLE IF NOT EXISTS spending_records (
    id            UUID          PRIMARY KEY,
    budget_id     UUID          NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
    allocation_id UUID          REFERENCES budget_allocations(id) ON DELETE SET NULL,
    category      VARCHAR(255)  NOT NULL DEFAULT '',
    description   TEXT          NOT NULL DEFAULT '',
    amount        NUMERIC(18,4) NOT NULL CHECK (amount > 0),
    reference     VARCHAR(255)  NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_spending_records_budget_id
    ON spending_records (budget_id);

CREATE INDEX IF NOT EXISTS idx_spending_records_created_at
    ON spending_records (created_at DESC);
