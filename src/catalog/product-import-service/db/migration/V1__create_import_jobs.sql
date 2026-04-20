CREATE TABLE IF NOT EXISTS import_jobs (
    id             TEXT         PRIMARY KEY,
    file_name      TEXT         NOT NULL,
    format         TEXT         NOT NULL,
    status         TEXT         NOT NULL DEFAULT 'PENDING',
    total_rows     INT          NOT NULL DEFAULT 0,
    processed_rows INT          NOT NULL DEFAULT 0,
    error_rows     INT          NOT NULL DEFAULT 0,
    errors         JSONB        NOT NULL DEFAULT '[]',
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    completed_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_import_jobs_status ON import_jobs(status);
CREATE INDEX IF NOT EXISTS idx_import_jobs_created_at ON import_jobs(created_at DESC);
