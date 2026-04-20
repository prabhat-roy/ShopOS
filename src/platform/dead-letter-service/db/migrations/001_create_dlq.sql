-- migrate:up
CREATE TABLE dead_messages (
    id           TEXT        PRIMARY KEY,
    topic        TEXT        NOT NULL,
    partition    INT         NOT NULL DEFAULT 0,
    "offset"     BIGINT      NOT NULL DEFAULT 0,
    key          TEXT        NOT NULL DEFAULT '',
    payload      JSONB       NOT NULL DEFAULT '{}',
    error_reason TEXT        NOT NULL DEFAULT '',
    status       TEXT        NOT NULL DEFAULT 'pending'
                             CHECK (status IN ('pending', 'retried', 'discarded')),
    retry_count  INT         NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_dlq_topic_status ON dead_messages (topic, status);
CREATE INDEX idx_dlq_created_at   ON dead_messages (created_at);

-- migrate:down
DROP TABLE IF EXISTS dead_messages;
