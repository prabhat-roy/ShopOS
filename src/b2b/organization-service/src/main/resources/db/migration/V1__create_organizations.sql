-- ─────────────────────────────────────────────────────────────────────────────
-- V1__create_organizations.sql
-- Creates organizations and org_members tables for the organization-service
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS organizations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(255) NOT NULL,
    slug            VARCHAR(255) NOT NULL,
    email           VARCHAR(320) NOT NULL,
    phone           VARCHAR(50),
    website         VARCHAR(2048),
    type            VARCHAR(50)  NOT NULL DEFAULT 'SMB',
    status          VARCHAR(50)  NOT NULL DEFAULT 'PENDING_VERIFICATION',
    industry        VARCHAR(255),
    tax_id          VARCHAR(100),
    country         VARCHAR(100),
    address         TEXT,
    employee_count  INT          NOT NULL DEFAULT 0,
    credit_limit    NUMERIC(19, 4) NOT NULL DEFAULT 0,
    parent_org_id   UUID,
    settings        TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),

    CONSTRAINT uq_organizations_slug  UNIQUE (slug),
    CONSTRAINT uq_organizations_email UNIQUE (email),
    CONSTRAINT fk_organizations_parent FOREIGN KEY (parent_org_id)
        REFERENCES organizations (id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_organizations_status       ON organizations (status);
CREATE INDEX IF NOT EXISTS idx_organizations_type         ON organizations (type);
CREATE INDEX IF NOT EXISTS idx_organizations_parent_org   ON organizations (parent_org_id);
CREATE INDEX IF NOT EXISTS idx_organizations_created_at   ON organizations (created_at DESC);

-- ─────────────────────────────────────────────────────────────────────────────
-- Org Members
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS org_members (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id       UUID         NOT NULL,
    user_id      UUID         NOT NULL,
    role         VARCHAR(100) NOT NULL,
    department   VARCHAR(255),
    job_title    VARCHAR(255),
    active       BOOLEAN      NOT NULL DEFAULT TRUE,
    invited_at   TIMESTAMPTZ,
    joined_at    TIMESTAMPTZ,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),

    CONSTRAINT uq_org_members_org_user UNIQUE (org_id, user_id),
    CONSTRAINT fk_org_members_org FOREIGN KEY (org_id)
        REFERENCES organizations (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_org_members_org_id  ON org_members (org_id);
CREATE INDEX IF NOT EXISTS idx_org_members_user_id ON org_members (user_id);
CREATE INDEX IF NOT EXISTS idx_org_members_active  ON org_members (org_id, active);
