-- gift_cards: one row per issued card
CREATE TABLE IF NOT EXISTS gift_cards (
    id              TEXT           NOT NULL PRIMARY KEY,
    code            TEXT           NOT NULL UNIQUE,
    initial_balance NUMERIC(18, 4) NOT NULL CHECK (initial_balance > 0),
    current_balance NUMERIC(18, 4) NOT NULL DEFAULT 0 CHECK (current_balance >= 0),
    currency        TEXT           NOT NULL DEFAULT 'USD',
    issued_to       TEXT           NOT NULL DEFAULT '',
    active          BOOLEAN        NOT NULL DEFAULT TRUE,
    expires_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_gift_cards_code ON gift_cards(code);

-- redemption_records: append-only log of redemptions
CREATE TABLE IF NOT EXISTS redemption_records (
    id         TEXT           NOT NULL PRIMARY KEY,
    card_id    TEXT           NOT NULL REFERENCES gift_cards(id),
    order_id   TEXT           NOT NULL DEFAULT '',
    amount     NUMERIC(18, 4) NOT NULL CHECK (amount > 0),
    created_at TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_redemption_records_card_id    ON redemption_records(card_id);
CREATE INDEX IF NOT EXISTS idx_redemption_records_created_at ON redemption_records(created_at DESC);
