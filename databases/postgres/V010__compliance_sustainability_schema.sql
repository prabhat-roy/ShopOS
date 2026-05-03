-- Compliance + sustainability — data retention, privacy requests, lineage, carbon, eco scores.

CREATE SCHEMA IF NOT EXISTS compliance AUTHORIZATION compliance_app;
SET search_path TO compliance, public;

CREATE TABLE retention_policy (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    data_class      VARCHAR(64) NOT NULL,
    retain_days     INT NOT NULL,
    legal_basis     VARCHAR(64),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE consent_audit (
    id              BIGSERIAL PRIMARY KEY,
    user_id         UUID NOT NULL,
    purpose         VARCHAR(64) NOT NULL,
    action          VARCHAR(16) NOT NULL CHECK (action IN ('granted','revoked')),
    source_ip       INET,
    user_agent      TEXT,
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX consent_audit_user_idx ON consent_audit (user_id, occurred_at DESC);

CREATE TABLE privacy_request (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL,
    request_type    VARCHAR(24) NOT NULL CHECK (request_type IN ('access','export','delete','rectify')),
    status          VARCHAR(24) NOT NULL DEFAULT 'received',
    received_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    due_at          TIMESTAMPTZ NOT NULL,
    completed_at    TIMESTAMPTZ
);

CREATE TABLE compliance_report (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    framework       VARCHAR(32) NOT NULL,
    period_start    DATE NOT NULL,
    period_end      DATE NOT NULL,
    pdf_uri         TEXT,
    generated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE data_lineage_node (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    namespace       VARCHAR(64) NOT NULL,
    name            VARCHAR(128) NOT NULL,
    node_type       VARCHAR(32) NOT NULL,
    UNIQUE (namespace, name)
);

CREATE TABLE data_lineage_edge (
    upstream_id     UUID NOT NULL REFERENCES data_lineage_node(id) ON DELETE CASCADE,
    downstream_id   UUID NOT NULL REFERENCES data_lineage_node(id) ON DELETE CASCADE,
    job             VARCHAR(128),
    PRIMARY KEY (upstream_id, downstream_id)
);

CREATE SCHEMA IF NOT EXISTS sustainability AUTHORIZATION sustainability_app;
SET search_path TO sustainability, public;

CREATE TABLE carbon_factor (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity        VARCHAR(64) NOT NULL,
    region          VARCHAR(16) NOT NULL,
    kg_co2_per_unit NUMERIC(10,4) NOT NULL,
    unit            VARCHAR(16) NOT NULL,
    source          VARCHAR(120),
    valid_from      DATE NOT NULL,
    UNIQUE (activity, region, valid_from)
);

CREATE TABLE order_carbon (
    order_id        UUID PRIMARY KEY,
    kg_co2          NUMERIC(10,4) NOT NULL,
    breakdown       JSONB NOT NULL,
    calculated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE eco_score (
    sku             VARCHAR(64) PRIMARY KEY,
    score           CHAR(1) NOT NULL CHECK (score IN ('A','B','C','D','E')),
    factors         JSONB NOT NULL,
    last_calculated TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE carbon_offset (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        UUID,
    provider        VARCHAR(64) NOT NULL,
    kg_co2          NUMERIC(10,4) NOT NULL,
    cost_cents      BIGINT NOT NULL,
    currency        CHAR(3) NOT NULL,
    purchased_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
