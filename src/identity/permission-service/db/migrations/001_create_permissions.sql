-- migrate:up

-- Roles table: stores named roles with an array of permission strings.
CREATE TABLE IF NOT EXISTS roles (
    id          TEXT        PRIMARY KEY,
    name        TEXT        NOT NULL UNIQUE,
    description TEXT        NOT NULL DEFAULT '',
    permissions TEXT[]      NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- User-role binding table: many-to-many between users and roles.
CREATE TABLE IF NOT EXISTS user_roles (
    user_id     TEXT        NOT NULL,
    role_id     TEXT        NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);

-- migrate:down

DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS roles;
