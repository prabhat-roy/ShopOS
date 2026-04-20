-- migrate:up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE feature_flags (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key         TEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    enabled     BOOLEAN NOT NULL DEFAULT FALSE,
    strategy    TEXT NOT NULL DEFAULT 'all'
                    CHECK (strategy IN ('all','percentage','user_list','context')),
    percentage  INT NOT NULL DEFAULT 0 CHECK (percentage >= 0 AND percentage <= 100),
    user_ids    TEXT[] NOT NULL DEFAULT '{}',
    context_key TEXT NOT NULL DEFAULT '',
    context_val TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_feature_flags_key ON feature_flags (key);

-- migrate:down
DROP TABLE IF EXISTS feature_flags;
