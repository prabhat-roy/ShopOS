-- migrate:up
CREATE TABLE tenants (
    id          TEXT        PRIMARY KEY,
    name        TEXT        NOT NULL,
    slug        TEXT        NOT NULL UNIQUE,
    plan        TEXT        NOT NULL DEFAULT 'starter',
    status      TEXT        NOT NULL DEFAULT 'active',
    owner_email TEXT        NOT NULL,
    settings    JSONB       NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tenants_slug   ON tenants(slug);
CREATE INDEX idx_tenants_status ON tenants(status);

-- migrate:down
DROP TABLE IF EXISTS tenants;
