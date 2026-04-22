-- Flyway migration V003 — Financial domain schema
-- Services: invoice-service, accounting-service, payout-service, reconciliation-service,
--           tax-reporting-service, credit-service, kyc-aml-service, escrow-service

-- ─── Invoices ────────────────────────────────────────────────────────────────
CREATE TABLE invoices (
    id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    invoice_number  TEXT        NOT NULL UNIQUE,
    order_id        UUID        NOT NULL,
    user_id         UUID        NOT NULL,
    status          TEXT        NOT NULL DEFAULT 'draft'
                    CHECK (status IN ('draft','issued','paid','overdue','cancelled','voided')),
    subtotal        NUMERIC(12,2) NOT NULL,
    tax_amount      NUMERIC(12,2) NOT NULL DEFAULT 0,
    total_amount    NUMERIC(12,2) NOT NULL,
    currency        CHAR(3)     NOT NULL DEFAULT 'USD',
    due_date        DATE,
    paid_at         TIMESTAMPTZ,
    billing_address JSONB       NOT NULL DEFAULT '{}',
    line_items      JSONB       NOT NULL DEFAULT '[]',
    notes           TEXT,
    metadata        JSONB       NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_invoices_order_id  ON invoices (order_id);
CREATE INDEX idx_invoices_user_id   ON invoices (user_id, created_at DESC);
CREATE INDEX idx_invoices_status    ON invoices (status, due_date);

-- ─── Payouts ──────────────────────────────────────────────────────────────────
CREATE TABLE payouts (
    id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    recipient_id    UUID        NOT NULL,
    recipient_type  TEXT        NOT NULL CHECK (recipient_type IN ('seller','affiliate','partner')),
    amount          NUMERIC(12,2) NOT NULL CHECK (amount > 0),
    currency        CHAR(3)     NOT NULL DEFAULT 'USD',
    status          TEXT        NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending','processing','completed','failed','cancelled')),
    method          TEXT        NOT NULL,
    reference       TEXT,
    failure_reason  TEXT,
    period_from     DATE        NOT NULL,
    period_to       DATE        NOT NULL,
    metadata        JSONB       NOT NULL DEFAULT '{}',
    scheduled_at    TIMESTAMPTZ,
    processed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payouts_recipient  ON payouts (recipient_id, created_at DESC);
CREATE INDEX idx_payouts_status     ON payouts (status, scheduled_at);

-- ─── GL Accounts (accounting) ─────────────────────────────────────────────────
CREATE TABLE gl_accounts (
    id          UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    code        TEXT    NOT NULL UNIQUE,
    name        TEXT    NOT NULL,
    type        TEXT    NOT NULL CHECK (type IN ('asset','liability','equity','revenue','expense')),
    currency    CHAR(3) NOT NULL DEFAULT 'USD',
    balance     NUMERIC(14,2) NOT NULL DEFAULT 0,
    parent_id   UUID    REFERENCES gl_accounts(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE journal_entries (
    id          BIGSERIAL   PRIMARY KEY,
    reference   TEXT        NOT NULL,
    description TEXT        NOT NULL,
    posted_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by  TEXT        NOT NULL
);

CREATE TABLE journal_lines (
    id              BIGSERIAL   PRIMARY KEY,
    journal_id      BIGINT      NOT NULL REFERENCES journal_entries(id),
    account_id      UUID        NOT NULL REFERENCES gl_accounts(id),
    debit           NUMERIC(12,2) NOT NULL DEFAULT 0,
    credit          NUMERIC(12,2) NOT NULL DEFAULT 0,
    currency        CHAR(3)     NOT NULL DEFAULT 'USD',
    description     TEXT,
    CONSTRAINT debit_or_credit CHECK (
        (debit > 0 AND credit = 0) OR (credit > 0 AND debit = 0)
    )
);

CREATE INDEX idx_journal_lines_account ON journal_lines (account_id);

-- ─── Escrow ───────────────────────────────────────────────────────────────────
CREATE TABLE escrow_accounts (
    id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id        UUID        NOT NULL UNIQUE,
    buyer_id        UUID        NOT NULL,
    seller_id       UUID        NOT NULL,
    amount          NUMERIC(12,2) NOT NULL,
    currency        CHAR(3)     NOT NULL DEFAULT 'USD',
    status          TEXT        NOT NULL DEFAULT 'holding'
                    CHECK (status IN ('holding','releasing','released','refunded','disputed')),
    release_at      TIMESTAMPTZ,
    released_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── KYC/AML ──────────────────────────────────────────────────────────────────
CREATE TABLE kyc_verifications (
    id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID        NOT NULL,
    level           TEXT        NOT NULL CHECK (level IN ('basic','enhanced','full')),
    status          TEXT        NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending','in_review','approved','rejected','expired')),
    provider        TEXT,
    provider_ref    TEXT,
    documents       JSONB       NOT NULL DEFAULT '[]',
    risk_score      NUMERIC(5,2),
    flags           TEXT[]      DEFAULT '{}',
    reviewed_by     TEXT,
    reviewed_at     TIMESTAMPTZ,
    expires_at      DATE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_kyc_user_id ON kyc_verifications (user_id, created_at DESC);

-- ─── Triggers ─────────────────────────────────────────────────────────────────
CREATE TRIGGER invoices_updated_at BEFORE UPDATE ON invoices FOR EACH ROW EXECUTE FUNCTION set_updated_at();
