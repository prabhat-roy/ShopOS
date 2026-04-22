-- Flyway migration V001 — Identity domain schema
-- Services: user-service, auth-service, session-service, permission-service, mfa-service

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ─── Users ────────────────────────────────────────────────────────────────────
CREATE TABLE users (
    id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    email           TEXT        NOT NULL UNIQUE,
    email_verified  BOOLEAN     NOT NULL DEFAULT FALSE,
    password_hash   TEXT,
    full_name       TEXT        NOT NULL,
    phone           TEXT,
    phone_verified  BOOLEAN     NOT NULL DEFAULT FALSE,
    avatar_url      TEXT,
    status          TEXT        NOT NULL DEFAULT 'active'
                    CHECK (status IN ('active', 'suspended', 'deleted', 'pending_verification')),
    locale          TEXT        NOT NULL DEFAULT 'en',
    timezone        TEXT        NOT NULL DEFAULT 'UTC',
    metadata        JSONB       NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_users_email         ON users (email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_status        ON users (status) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_created_at    ON users (created_at DESC);

-- ─── Roles ────────────────────────────────────────────────────────────────────
CREATE TABLE roles (
    id          UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        TEXT    NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO roles (name, description) VALUES
    ('CUSTOMER',   'Regular storefront customer'),
    ('SELLER',     'Marketplace seller'),
    ('ADMIN',      'Platform administrator'),
    ('SUPERADMIN', 'Full platform access'),
    ('PARTNER',    'B2B partner user'),
    ('DEVELOPER',  'API developer');

CREATE TABLE user_roles (
    user_id     UUID REFERENCES users(id) ON DELETE CASCADE,
    role_id     UUID REFERENCES roles(id) ON DELETE CASCADE,
    granted_by  UUID REFERENCES users(id),
    granted_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

-- ─── Sessions ─────────────────────────────────────────────────────────────────
CREATE TABLE sessions (
    id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token   TEXT        NOT NULL UNIQUE,
    device_id       TEXT,
    ip_address      INET,
    user_agent      TEXT,
    expires_at      TIMESTAMPTZ NOT NULL,
    revoked_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sessions_user_id       ON sessions (user_id);
CREATE INDEX idx_sessions_refresh_token ON sessions (refresh_token) WHERE revoked_at IS NULL;

-- ─── MFA ──────────────────────────────────────────────────────────────────────
CREATE TABLE mfa_configs (
    id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID        NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    type        TEXT        NOT NULL CHECK (type IN ('totp', 'sms', 'email')),
    secret      TEXT,
    enabled     BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE mfa_backup_codes (
    id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash   TEXT        NOT NULL,
    used_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── API Keys ─────────────────────────────────────────────────────────────────
CREATE TABLE api_keys (
    id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT        NOT NULL,
    key_hash    TEXT        NOT NULL UNIQUE,
    key_prefix  TEXT        NOT NULL,
    scopes      TEXT[]      NOT NULL DEFAULT '{}',
    environment TEXT        NOT NULL DEFAULT 'production'
                CHECK (environment IN ('sandbox', 'production')),
    last_used_at TIMESTAMPTZ,
    expires_at  TIMESTAMPTZ,
    revoked_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_api_keys_user_id  ON api_keys (user_id) WHERE revoked_at IS NULL;
CREATE INDEX idx_api_keys_key_hash ON api_keys (key_hash) WHERE revoked_at IS NULL;

-- ─── Audit log ────────────────────────────────────────────────────────────────
CREATE TABLE auth_audit_log (
    id          BIGSERIAL   PRIMARY KEY,
    user_id     UUID        REFERENCES users(id),
    event_type  TEXT        NOT NULL,
    ip_address  INET,
    user_agent  TEXT,
    metadata    JSONB       NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auth_audit_user    ON auth_audit_log (user_id, created_at DESC);
CREATE INDEX idx_auth_audit_event   ON auth_audit_log (event_type, created_at DESC);

-- ─── Updated_at trigger ───────────────────────────────────────────────────────
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN NEW.updated_at = NOW(); RETURN NEW; END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER users_updated_at    BEFORE UPDATE ON users    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER mfa_updated_at      BEFORE UPDATE ON mfa_configs FOR EACH ROW EXECUTE FUNCTION set_updated_at();
