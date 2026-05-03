-- B2B — organizations, contracts, quotes, RFx, approvals, credit limits.

CREATE SCHEMA IF NOT EXISTS b2b AUTHORIZATION b2b_app;
SET search_path TO b2b, public;

CREATE TABLE organization (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    legal_name      VARCHAR(255) NOT NULL,
    duns            VARCHAR(9) UNIQUE,
    tax_id          VARCHAR(64),
    country_iso2    CHAR(2) NOT NULL,
    industry_code   VARCHAR(8),
    parent_org_id   UUID REFERENCES organization(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE org_member (
    org_id          UUID NOT NULL REFERENCES organization(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL,
    role            VARCHAR(32) NOT NULL CHECK (role IN ('admin','buyer','approver','viewer')),
    PRIMARY KEY (org_id, user_id)
);

CREATE TABLE contract (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organization(id),
    contract_no     VARCHAR(48) NOT NULL UNIQUE,
    starts_on       DATE NOT NULL,
    ends_on         DATE NOT NULL,
    payment_terms   VARCHAR(32),
    discount_pct    NUMERIC(5,2) NOT NULL DEFAULT 0,
    pdf_uri         TEXT,
    status          VARCHAR(24) NOT NULL DEFAULT 'active'
);

CREATE TABLE quote (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organization(id),
    quote_no        VARCHAR(32) NOT NULL UNIQUE,
    status          VARCHAR(24) NOT NULL DEFAULT 'draft',
    total_cents     BIGINT NOT NULL DEFAULT 0,
    currency        CHAR(3) NOT NULL,
    expires_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE quote_line (
    quote_id        UUID NOT NULL REFERENCES quote(id) ON DELETE CASCADE,
    line_no         INT NOT NULL,
    sku             VARCHAR(64) NOT NULL,
    qty             INT NOT NULL CHECK (qty > 0),
    unit_cents      BIGINT NOT NULL,
    PRIMARY KEY (quote_id, line_no)
);

CREATE TABLE rfp (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organization(id),
    title           VARCHAR(255) NOT NULL,
    description     TEXT,
    closes_at       TIMESTAMPTZ NOT NULL,
    status          VARCHAR(24) NOT NULL DEFAULT 'open',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE approval_workflow (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type     VARCHAR(32) NOT NULL,
    entity_id       UUID NOT NULL,
    requested_by    UUID NOT NULL,
    current_step    INT NOT NULL DEFAULT 1,
    total_steps     INT NOT NULL,
    status          VARCHAR(24) NOT NULL DEFAULT 'pending',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE credit_limit (
    org_id          UUID PRIMARY KEY REFERENCES organization(id) ON DELETE CASCADE,
    limit_cents     BIGINT NOT NULL DEFAULT 0,
    used_cents      BIGINT NOT NULL DEFAULT 0,
    currency        CHAR(3) NOT NULL,
    review_at       DATE
);

CREATE TABLE purchase_requisition (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organization(id),
    requested_by    UUID NOT NULL,
    cost_center     VARCHAR(32),
    total_cents     BIGINT NOT NULL DEFAULT 0,
    status          VARCHAR(24) NOT NULL DEFAULT 'draft',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
