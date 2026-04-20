-- V1: Initial schema for ad-service
-- Creates ad_campaigns and ad_impressions tables

CREATE TABLE IF NOT EXISTS ad_campaigns (
    id                 UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name               VARCHAR(255) NOT NULL,
    advertiser_id      UUID         NOT NULL,
    status             VARCHAR(20)  NOT NULL DEFAULT 'DRAFT',
    ad_type            VARCHAR(30)  NOT NULL,
    target_categories  TEXT,
    target_audience    TEXT,
    budget             NUMERIC(19, 4) NOT NULL,
    spent              NUMERIC(19, 4) NOT NULL DEFAULT 0,
    impressions        BIGINT       NOT NULL DEFAULT 0,
    clicks             BIGINT       NOT NULL DEFAULT 0,
    start_date         DATE         NOT NULL,
    end_date           DATE         NOT NULL,
    image_url          VARCHAR(2048),
    target_url         VARCHAR(2048),
    bid_amount         NUMERIC(19, 4),
    created_at         TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMP    NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_status   CHECK (status  IN ('DRAFT', 'ACTIVE', 'PAUSED', 'COMPLETED', 'CANCELLED')),
    CONSTRAINT chk_ad_type  CHECK (ad_type IN ('BANNER', 'SPONSORED_PRODUCT', 'POPUP', 'NATIVE', 'VIDEO')),
    CONSTRAINT chk_dates    CHECK (end_date >= start_date),
    CONSTRAINT chk_budget   CHECK (budget > 0),
    CONSTRAINT chk_spent    CHECK (spent >= 0),
    CONSTRAINT chk_impr     CHECK (impressions >= 0),
    CONSTRAINT chk_clicks   CHECK (clicks >= 0)
);

CREATE INDEX IF NOT EXISTS idx_campaign_advertiser_id ON ad_campaigns (advertiser_id);
CREATE INDEX IF NOT EXISTS idx_campaign_status        ON ad_campaigns (status);
CREATE INDEX IF NOT EXISTS idx_campaign_dates         ON ad_campaigns (start_date, end_date);
CREATE INDEX IF NOT EXISTS idx_campaign_type_status   ON ad_campaigns (ad_type, status);

CREATE TABLE IF NOT EXISTS ad_impressions (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id  UUID         NOT NULL REFERENCES ad_campaigns(id),
    user_id      VARCHAR(255),
    session_id   VARCHAR(255) NOT NULL,
    placement_id VARCHAR(255) NOT NULL,
    shown_at     TIMESTAMP    NOT NULL DEFAULT NOW(),
    clicked      BOOLEAN      NOT NULL DEFAULT FALSE,
    clicked_at   TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_impression_campaign_id ON ad_impressions (campaign_id);
CREATE INDEX IF NOT EXISTS idx_impression_session_id  ON ad_impressions (session_id);
CREATE INDEX IF NOT EXISTS idx_impression_shown_at    ON ad_impressions (shown_at);
