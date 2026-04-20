-- Migration: 001_init.up.sql
-- Creates the expenses table for the expense-management-service.

CREATE TABLE IF NOT EXISTS expenses (
    id              UUID        NOT NULL PRIMARY KEY,
    employee_id     UUID        NOT NULL,
    category        TEXT        NOT NULL
        CONSTRAINT chk_expenses_category
            CHECK (category IN ('TRAVEL','MEALS','SOFTWARE','HARDWARE','OFFICE','MARKETING','OTHER')),
    amount          NUMERIC(18, 4) NOT NULL
        CONSTRAINT chk_expenses_amount_positive CHECK (amount > 0),
    currency        CHAR(3)     NOT NULL,
    description     TEXT        NOT NULL,
    receipt_url     TEXT        NOT NULL DEFAULT '',
    status          TEXT        NOT NULL DEFAULT 'DRAFT'
        CONSTRAINT chk_expenses_status
            CHECK (status IN ('DRAFT','SUBMITTED','APPROVED','REJECTED','REIMBURSED')),
    approved_by     UUID,
    approved_at     TIMESTAMPTZ,
    reimbursed_at   TIMESTAMPTZ,
    notes           TEXT        NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- An expense can only carry an approver once it is APPROVED.
    CONSTRAINT chk_expenses_approved_by
        CHECK (
            (status NOT IN ('APPROVED','REIMBURSED')) OR (approved_by IS NOT NULL)
        ),
    -- reimbursed_at must be set when REIMBURSED.
    CONSTRAINT chk_expenses_reimbursed_at
        CHECK (
            status != 'REIMBURSED' OR reimbursed_at IS NOT NULL
        )
);

-- Index for employee lookups.
CREATE INDEX IF NOT EXISTS idx_expenses_employee_id
    ON expenses (employee_id);

-- Index for status-based filtering (e.g., pending approvals).
CREATE INDEX IF NOT EXISTS idx_expenses_status
    ON expenses (status);

-- Index for category reporting.
CREATE INDEX IF NOT EXISTS idx_expenses_category
    ON expenses (category);

-- Composite index for employee + status queries.
CREATE INDEX IF NOT EXISTS idx_expenses_employee_status
    ON expenses (employee_id, status);

-- Index for time-ordered listing.
CREATE INDEX IF NOT EXISTS idx_expenses_created_at
    ON expenses (created_at DESC);
