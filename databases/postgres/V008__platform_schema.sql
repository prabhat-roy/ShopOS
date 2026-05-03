-- Platform — saga state, event store, webhooks, schedules, idempotency, feature flags, tenants.

CREATE SCHEMA IF NOT EXISTS platform AUTHORIZATION platform_app;
SET search_path TO platform, public;

CREATE TABLE saga_instance (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    saga_type       VARCHAR(64) NOT NULL,
    correlation_id  VARCHAR(128) NOT NULL,
    status          VARCHAR(24) NOT NULL DEFAULT 'started',
    current_step    VARCHAR(64) NOT NULL,
    payload         JSONB NOT NULL DEFAULT '{}'::jsonb,
    started_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at    TIMESTAMPTZ,
    UNIQUE (saga_type, correlation_id)
);
CREATE INDEX saga_status_idx ON saga_instance (status) WHERE status NOT IN ('completed','compensated');

CREATE TABLE saga_step (
    saga_id         UUID NOT NULL REFERENCES saga_instance(id) ON DELETE CASCADE,
    step_no         INT NOT NULL,
    name            VARCHAR(64) NOT NULL,
    status          VARCHAR(24) NOT NULL,
    attempts        INT NOT NULL DEFAULT 0,
    last_error      TEXT,
    executed_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (saga_id, step_no)
);

CREATE TABLE event_store (
    id              BIGSERIAL PRIMARY KEY,
    aggregate_type  VARCHAR(64) NOT NULL,
    aggregate_id    UUID NOT NULL,
    sequence_no     BIGINT NOT NULL,
    event_type      VARCHAR(128) NOT NULL,
    payload         JSONB NOT NULL,
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (aggregate_id, sequence_no)
);
CREATE INDEX event_store_aggregate_idx ON event_store (aggregate_type, aggregate_id, sequence_no);

CREATE TABLE webhook_subscription (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    event_pattern   VARCHAR(128) NOT NULL,
    target_url      TEXT NOT NULL,
    secret          VARCHAR(64) NOT NULL,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE webhook_delivery (
    id              BIGSERIAL PRIMARY KEY,
    subscription_id UUID NOT NULL REFERENCES webhook_subscription(id) ON DELETE CASCADE,
    event_id        BIGINT NOT NULL,
    attempt         INT NOT NULL DEFAULT 1,
    response_code   INT,
    response_body   TEXT,
    delivered_at    TIMESTAMPTZ
);
CREATE INDEX webhook_pending_idx ON webhook_delivery (subscription_id) WHERE delivered_at IS NULL;

CREATE TABLE scheduled_job (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(120) NOT NULL UNIQUE,
    cron            VARCHAR(64) NOT NULL,
    job_class       VARCHAR(160) NOT NULL,
    payload         JSONB NOT NULL DEFAULT '{}'::jsonb,
    last_run_at     TIMESTAMPTZ,
    next_run_at     TIMESTAMPTZ NOT NULL,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE feature_flag (
    key             VARCHAR(120) PRIMARY KEY,
    description     TEXT,
    enabled         BOOLEAN NOT NULL DEFAULT FALSE,
    rules           JSONB NOT NULL DEFAULT '[]'::jsonb,
    rollout_pct     NUMERIC(5,2) NOT NULL DEFAULT 0,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by      VARCHAR(120)
);

CREATE TABLE tenant (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug            VARCHAR(64) NOT NULL UNIQUE,
    name            VARCHAR(160) NOT NULL,
    plan            VARCHAR(32) NOT NULL DEFAULT 'standard',
    region          VARCHAR(16) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    suspended_at    TIMESTAMPTZ
);

CREATE TABLE api_versioning (
    id              SERIAL PRIMARY KEY,
    api_name        VARCHAR(64) NOT NULL,
    version         VARCHAR(16) NOT NULL,
    deprecation_at  TIMESTAMPTZ,
    sunset_at       TIMESTAMPTZ,
    UNIQUE (api_name, version)
);
