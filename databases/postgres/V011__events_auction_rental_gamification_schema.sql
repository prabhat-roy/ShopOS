-- Events/Ticketing, Auction, Rental, Gamification — relational data for these new domains.

CREATE SCHEMA IF NOT EXISTS events_ticketing AUTHORIZATION events_ticketing_app;
SET search_path TO events_ticketing, public;

CREATE TABLE venue (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(160) NOT NULL,
    address         TEXT,
    city            VARCHAR(120),
    country_iso2    CHAR(2) NOT NULL,
    capacity        INT NOT NULL CHECK (capacity > 0),
    timezone        VARCHAR(64) NOT NULL
);

CREATE TABLE event (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    venue_id        UUID NOT NULL REFERENCES venue(id),
    title           VARCHAR(255) NOT NULL,
    starts_at       TIMESTAMPTZ NOT NULL,
    ends_at         TIMESTAMPTZ,
    status          VARCHAR(24) NOT NULL DEFAULT 'scheduled',
    on_sale_at      TIMESTAMPTZ
);

CREATE TABLE seat_map (
    venue_id        UUID NOT NULL REFERENCES venue(id) ON DELETE CASCADE,
    section         VARCHAR(32) NOT NULL,
    row_label       VARCHAR(8) NOT NULL,
    seat_no         VARCHAR(8) NOT NULL,
    PRIMARY KEY (venue_id, section, row_label, seat_no)
);

CREATE TABLE ticket (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id        UUID NOT NULL REFERENCES event(id) ON DELETE CASCADE,
    section         VARCHAR(32),
    row_label       VARCHAR(8),
    seat_no         VARCHAR(8),
    price_cents     BIGINT NOT NULL,
    currency        CHAR(3) NOT NULL,
    status          VARCHAR(24) NOT NULL DEFAULT 'available',
    holder_user_id  UUID,
    UNIQUE (event_id, section, row_label, seat_no)
);

CREATE TABLE booking (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL,
    event_id        UUID NOT NULL REFERENCES event(id),
    total_cents     BIGINT NOT NULL,
    currency        CHAR(3) NOT NULL,
    status          VARCHAR(24) NOT NULL DEFAULT 'pending',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE check_in (
    ticket_id       UUID PRIMARY KEY REFERENCES ticket(id),
    checked_in_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    gate            VARCHAR(16),
    operator        VARCHAR(64)
);

-- Auction
CREATE SCHEMA IF NOT EXISTS auction AUTHORIZATION auction_app;
SET search_path TO auction, public;

CREATE TABLE auction (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sku             VARCHAR(64),
    starts_at       TIMESTAMPTZ NOT NULL,
    ends_at         TIMESTAMPTZ NOT NULL,
    starting_cents  BIGINT NOT NULL,
    reserve_cents   BIGINT,
    bid_increment   BIGINT NOT NULL DEFAULT 100,
    currency        CHAR(3) NOT NULL,
    status          VARCHAR(24) NOT NULL DEFAULT 'pending'
);

CREATE TABLE bid (
    id              BIGSERIAL PRIMARY KEY,
    auction_id      UUID NOT NULL REFERENCES auction(id) ON DELETE CASCADE,
    bidder_user_id  UUID NOT NULL,
    amount_cents    BIGINT NOT NULL,
    placed_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_proxy        BOOLEAN NOT NULL DEFAULT FALSE,
    max_proxy_cents BIGINT
);
CREATE INDEX bid_auction_idx ON bid (auction_id, amount_cents DESC);

CREATE TABLE auction_settlement (
    auction_id      UUID PRIMARY KEY REFERENCES auction(id),
    winner_user_id  UUID NOT NULL,
    final_cents     BIGINT NOT NULL,
    fee_cents       BIGINT NOT NULL DEFAULT 0,
    settled_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Rental
CREATE SCHEMA IF NOT EXISTS rental AUTHORIZATION rental_app;
SET search_path TO rental, public;

CREATE TABLE rental_item (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sku             VARCHAR(64) NOT NULL UNIQUE,
    name            VARCHAR(160) NOT NULL,
    daily_cents     BIGINT NOT NULL,
    deposit_cents   BIGINT NOT NULL DEFAULT 0,
    currency        CHAR(3) NOT NULL,
    is_available    BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE lease (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL,
    rental_item_id  UUID NOT NULL REFERENCES rental_item(id),
    start_date      DATE NOT NULL,
    end_date        DATE NOT NULL,
    total_cents     BIGINT NOT NULL,
    deposit_held_cents BIGINT NOT NULL,
    status          VARCHAR(24) NOT NULL DEFAULT 'active',
    EXCLUDE USING gist (rental_item_id WITH =, daterange(start_date, end_date) WITH &&)
);

CREATE TABLE damage_claim (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lease_id        UUID NOT NULL REFERENCES lease(id),
    description     TEXT,
    deduction_cents BIGINT NOT NULL,
    photos          TEXT[],
    filed_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE availability_block (
    rental_item_id  UUID NOT NULL REFERENCES rental_item(id) ON DELETE CASCADE,
    block_start     DATE NOT NULL,
    block_end       DATE NOT NULL,
    reason          VARCHAR(64),
    PRIMARY KEY (rental_item_id, block_start)
);

-- Gamification
CREATE SCHEMA IF NOT EXISTS gamification AUTHORIZATION gamification_app;
SET search_path TO gamification, public;

CREATE TABLE badge (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code            VARCHAR(64) NOT NULL UNIQUE,
    name            VARCHAR(120) NOT NULL,
    description     TEXT,
    icon_url        TEXT,
    points          INT NOT NULL DEFAULT 0
);

CREATE TABLE user_badge (
    user_id         UUID NOT NULL,
    badge_id        UUID NOT NULL REFERENCES badge(id) ON DELETE CASCADE,
    awarded_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, badge_id)
);

CREATE TABLE challenge (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code            VARCHAR(64) NOT NULL UNIQUE,
    name            VARCHAR(120) NOT NULL,
    rules           JSONB NOT NULL,
    starts_at       TIMESTAMPTZ NOT NULL,
    ends_at         TIMESTAMPTZ NOT NULL,
    reward_points   INT NOT NULL DEFAULT 0
);

CREATE TABLE challenge_progress (
    challenge_id    UUID NOT NULL REFERENCES challenge(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL,
    progress        JSONB NOT NULL DEFAULT '{}'::jsonb,
    completed_at    TIMESTAMPTZ,
    PRIMARY KEY (challenge_id, user_id)
);
