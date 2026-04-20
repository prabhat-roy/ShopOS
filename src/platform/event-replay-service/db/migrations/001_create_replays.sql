-- migrate:up
CREATE TABLE replay_jobs (
  id              TEXT        PRIMARY KEY,
  stream_id       TEXT        NOT NULL DEFAULT '',
  stream_type     TEXT        NOT NULL DEFAULT '',
  event_type      TEXT        NOT NULL DEFAULT '',
  from_seq        BIGINT      NOT NULL DEFAULT 0,
  to_seq          BIGINT      NOT NULL DEFAULT 0,
  from_time       TIMESTAMPTZ,
  to_time         TIMESTAMPTZ,
  target          TEXT        NOT NULL DEFAULT 'http',
  target_topic    TEXT        NOT NULL DEFAULT '',
  status          TEXT        NOT NULL DEFAULT 'pending',
  events_replayed BIGINT      NOT NULL DEFAULT 0,
  error_message   TEXT        NOT NULL DEFAULT '',
  started_at      TIMESTAMPTZ,
  completed_at    TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_replays_status ON replay_jobs(status);

-- migrate:down
DROP TABLE IF EXISTS replay_jobs;
