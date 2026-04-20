-- ─────────────────────────────────────────────────────────────────────────────
-- V1__create_promotions.sql
-- ShopOS :: commerce :: promotions-service
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS promotions (
    id                UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    code              TEXT            NOT NULL,
    name              TEXT            NOT NULL,
    type              TEXT            NOT NULL,
    discount_value    NUMERIC(12, 2),
    discount_percent  NUMERIC(5, 2),
    min_order_amount  NUMERIC(12, 2)  NOT NULL DEFAULT 0.00,
    max_uses          INTEGER         NOT NULL DEFAULT 0,
    used_count        INTEGER         NOT NULL DEFAULT 0,
    active            BOOLEAN         NOT NULL DEFAULT TRUE,
    starts_at         TIMESTAMPTZ,
    expires_at        TIMESTAMPTZ,
    created_at        TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ     NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_promotions_code UNIQUE (code)
);

-- Index: direct lookup by code (most frequent query path)
CREATE INDEX IF NOT EXISTS idx_promotions_code
    ON promotions (code);

-- Index: active promotions within a validity window (listPromotions activeOnly)
CREATE INDEX IF NOT EXISTS idx_promotions_active_dates
    ON promotions (active, starts_at, expires_at)
    WHERE active = TRUE;

-- Auto-update updated_at on every row change
CREATE OR REPLACE FUNCTION update_promotions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_promotions_updated_at ON promotions;

CREATE TRIGGER trg_promotions_updated_at
    BEFORE UPDATE ON promotions
    FOR EACH ROW
    EXECUTE FUNCTION update_promotions_updated_at();
