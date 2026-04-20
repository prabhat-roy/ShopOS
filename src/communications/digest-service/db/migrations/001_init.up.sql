-- ─── digest_configs ──────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS digest_configs (
    id           UUID        PRIMARY KEY,
    user_id      UUID        NOT NULL,
    email        TEXT        NOT NULL,
    frequency    TEXT        NOT NULL CHECK (frequency IN ('DAILY','WEEKLY')),
    status       TEXT        NOT NULL CHECK (status IN ('ACTIVE','PAUSED')) DEFAULT 'ACTIVE',
    last_sent_at TIMESTAMPTZ,
    next_send_at TIMESTAMPTZ NOT NULL,
    timezone     TEXT        NOT NULL DEFAULT 'UTC',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Efficient lookup for the scheduler: ACTIVE configs whose next_send_at is due.
CREATE INDEX IF NOT EXISTS idx_digest_configs_due
    ON digest_configs (status, next_send_at)
    WHERE status = 'ACTIVE';

-- Fast per-user queries.
CREATE INDEX IF NOT EXISTS idx_digest_configs_user_id
    ON digest_configs (user_id);

-- ─── digest_runs ─────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS digest_runs (
    id          UUID        PRIMARY KEY,
    config_id   UUID        NOT NULL REFERENCES digest_configs(id) ON DELETE CASCADE,
    sent_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    item_count  INT         NOT NULL DEFAULT 0,
    status      TEXT        NOT NULL,
    error_msg   TEXT        NOT NULL DEFAULT ''
);

-- Fast lookup of runs per config, ordered by most recent.
CREATE INDEX IF NOT EXISTS idx_digest_runs_config_id_sent_at
    ON digest_runs (config_id, sent_at DESC);
