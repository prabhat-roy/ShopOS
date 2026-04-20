-- migrate:up
CREATE TABLE jobs (
    id          TEXT        PRIMARY KEY,
    name        TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    cron_expr   TEXT        NOT NULL,
    http_method TEXT        NOT NULL DEFAULT 'POST',
    http_url    TEXT        NOT NULL,
    http_body   TEXT        NOT NULL DEFAULT '',
    status      TEXT        NOT NULL DEFAULT 'enabled' CHECK (status IN ('enabled','disabled')),
    next_run_at TIMESTAMPTZ NOT NULL,
    last_run_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_jobs_status_next ON jobs (status, next_run_at);

CREATE TABLE job_runs (
    id          TEXT        PRIMARY KEY,
    job_id      TEXT        NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    status      TEXT        NOT NULL CHECK (status IN ('success','failed')),
    output      TEXT        NOT NULL DEFAULT '',
    started_at  TIMESTAMPTZ NOT NULL,
    finished_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_job_runs_job_id     ON job_runs (job_id, started_at DESC);
CREATE INDEX idx_job_runs_started_at ON job_runs (started_at);

-- migrate:down
DROP TABLE IF EXISTS job_runs;
DROP TABLE IF EXISTS jobs;
