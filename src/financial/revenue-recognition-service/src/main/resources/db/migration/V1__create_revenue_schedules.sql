-- V1__create_revenue_schedules.sql
-- ASC 606 / IFRS 15 revenue recognition schedule table

CREATE TABLE IF NOT EXISTS revenue_schedules (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id                VARCHAR(255) NOT NULL,
    line_item_id            VARCHAR(255) NOT NULL,
    contract_type           VARCHAR(30)  NOT NULL,
    total_amount            NUMERIC(19, 4) NOT NULL,
    recognized_amount       NUMERIC(19, 4) NOT NULL DEFAULT 0,
    deferred_amount         NUMERIC(19, 4) NOT NULL,
    currency                CHAR(3) NOT NULL,
    recognition_start_date  DATE NOT NULL,
    recognition_end_date    DATE NOT NULL,
    status                  VARCHAR(30) NOT NULL DEFAULT 'PENDING',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rev_sched_order_id   ON revenue_schedules (order_id);
CREATE INDEX IF NOT EXISTS idx_rev_sched_status     ON revenue_schedules (status);
CREATE INDEX IF NOT EXISTS idx_rev_sched_start_date ON revenue_schedules (recognition_start_date);
CREATE INDEX IF NOT EXISTS idx_rev_sched_end_date   ON revenue_schedules (recognition_end_date);
