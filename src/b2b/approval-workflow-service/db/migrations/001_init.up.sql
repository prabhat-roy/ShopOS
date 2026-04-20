-- 001_init.up.sql
-- Creates the approval_workflows table for the approval-workflow-service.

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS approval_workflows (
    id                  UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_id           UUID        NOT NULL,
    entity_type         TEXT        NOT NULL
                            CHECK (entity_type IN ('purchase_order','quote','contract','expense')),
    org_id              UUID        NOT NULL,
    total_amount        NUMERIC(18,4) NOT NULL DEFAULT 0,
    status              TEXT        NOT NULL DEFAULT 'PENDING'
                            CHECK (status IN ('PENDING','IN_PROGRESS','APPROVED','REJECTED','CANCELLED')),
    steps               JSONB       NOT NULL DEFAULT '[]',
    current_step_index  INT         NOT NULL DEFAULT 0,
    created_by          TEXT        NOT NULL,
    completed_at        TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_aw_entity_id   ON approval_workflows (entity_id);
CREATE INDEX IF NOT EXISTS idx_aw_org_id      ON approval_workflows (org_id);
CREATE INDEX IF NOT EXISTS idx_aw_status      ON approval_workflows (status);
CREATE INDEX IF NOT EXISTS idx_aw_entity_type ON approval_workflows (entity_type);
CREATE INDEX IF NOT EXISTS idx_aw_created_at  ON approval_workflows (created_at DESC);
