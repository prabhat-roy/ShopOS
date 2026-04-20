-- migrate:up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE events (
    id          UUID        PRIMARY KEY,
    stream_id   TEXT        NOT NULL,
    stream_type TEXT        NOT NULL,
    event_type  TEXT        NOT NULL,
    version     BIGINT      NOT NULL,
    global_seq  BIGSERIAL   NOT NULL,
    payload     JSONB       NOT NULL,
    metadata    JSONB       NOT NULL DEFAULT '{}',
    occurred_at TIMESTAMPTZ NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_stream_version UNIQUE (stream_id, version)
);

CREATE INDEX idx_events_stream_id   ON events (stream_id, version);
CREATE INDEX idx_events_global_seq  ON events (global_seq);
CREATE INDEX idx_events_stream_type ON events (stream_type);
CREATE INDEX idx_events_event_type  ON events (event_type);
CREATE INDEX idx_events_occurred_at ON events (occurred_at);

CREATE TABLE snapshots (
    stream_id   TEXT        PRIMARY KEY,
    stream_type TEXT        NOT NULL,
    version     BIGINT      NOT NULL,
    state       JSONB       NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- migrate:down
DROP TABLE IF EXISTS snapshots;
DROP TABLE IF EXISTS events;
