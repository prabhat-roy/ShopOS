-- Customer experience — wishlists, support tickets, surveys, gift registry, alerts.

CREATE SCHEMA IF NOT EXISTS cx AUTHORIZATION cx_app;
SET search_path TO cx, public;

CREATE TABLE wishlist (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL,
    name            VARCHAR(120) NOT NULL DEFAULT 'Default',
    is_public       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE wishlist_item (
    wishlist_id     UUID NOT NULL REFERENCES wishlist(id) ON DELETE CASCADE,
    sku             VARCHAR(64) NOT NULL,
    added_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (wishlist_id, sku)
);

CREATE TABLE support_ticket (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL,
    subject         VARCHAR(255) NOT NULL,
    priority        VARCHAR(16) NOT NULL DEFAULT 'normal',
    status          VARCHAR(24) NOT NULL DEFAULT 'open',
    assignee_id     UUID,
    opened_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    first_response_at TIMESTAMPTZ,
    resolved_at     TIMESTAMPTZ
);
CREATE INDEX ticket_open_idx ON support_ticket (priority, opened_at) WHERE status NOT IN ('resolved','closed');

CREATE TABLE ticket_message (
    id              BIGSERIAL PRIMARY KEY,
    ticket_id       UUID NOT NULL REFERENCES support_ticket(id) ON DELETE CASCADE,
    author_type     VARCHAR(16) NOT NULL CHECK (author_type IN ('customer','agent','system')),
    author_id       UUID,
    body            TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE consent_record (
    user_id         UUID NOT NULL,
    purpose         VARCHAR(64) NOT NULL,
    granted         BOOLEAN NOT NULL,
    version         VARCHAR(16) NOT NULL,
    granted_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at      TIMESTAMPTZ,
    PRIMARY KEY (user_id, purpose, version)
);

CREATE TABLE survey (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title           VARCHAR(255) NOT NULL,
    questions       JSONB NOT NULL,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE survey_response (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    survey_id       UUID NOT NULL REFERENCES survey(id) ON DELETE CASCADE,
    user_id         UUID,
    answers         JSONB NOT NULL,
    submitted_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE gift_registry (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id   UUID NOT NULL,
    occasion        VARCHAR(64),
    event_date      DATE,
    handle          VARCHAR(64) NOT NULL UNIQUE,
    is_public       BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE gift_registry_item (
    registry_id     UUID NOT NULL REFERENCES gift_registry(id) ON DELETE CASCADE,
    sku             VARCHAR(64) NOT NULL,
    qty_requested   INT NOT NULL DEFAULT 1,
    qty_purchased   INT NOT NULL DEFAULT 0,
    PRIMARY KEY (registry_id, sku)
);

CREATE TABLE feedback (
    id              BIGSERIAL PRIMARY KEY,
    user_id         UUID,
    page_url        TEXT,
    rating          INT CHECK (rating BETWEEN 1 AND 5),
    comment         TEXT,
    submitted_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
