-- Migration: 001_create_consents
-- Description: Create consents and consent_history tables for GDPR consent management

CREATE TABLE IF NOT EXISTS consents (
    customer_id   VARCHAR(255)  NOT NULL,
    type          VARCHAR(50)   NOT NULL,
    granted       BOOLEAN       NOT NULL DEFAULT false,
    source        VARCHAR(255)  NOT NULL,
    ip_address    VARCHAR(45),
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    CONSTRAINT consents_pkey PRIMARY KEY (customer_id, type),
    CONSTRAINT consents_type_check CHECK (
        type IN (
            'MARKETING_EMAIL',
            'MARKETING_SMS',
            'ANALYTICS',
            'PERSONALIZATION',
            'THIRD_PARTY_SHARING',
            'ESSENTIAL'
        )
    )
);

CREATE INDEX IF NOT EXISTS idx_consents_customer_id ON consents (customer_id);
CREATE INDEX IF NOT EXISTS idx_consents_type ON consents (type);
CREATE INDEX IF NOT EXISTS idx_consents_granted ON consents (granted);

CREATE TABLE IF NOT EXISTS consent_history (
    id            UUID          NOT NULL DEFAULT gen_random_uuid(),
    customer_id   VARCHAR(255)  NOT NULL,
    type          VARCHAR(50)   NOT NULL,
    action        VARCHAR(10)   NOT NULL,
    source        VARCHAR(255),
    ip_address    VARCHAR(45),
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    CONSTRAINT consent_history_pkey PRIMARY KEY (id),
    CONSTRAINT consent_history_action_check CHECK (action IN ('grant', 'revoke')),
    CONSTRAINT consent_history_type_check CHECK (
        type IN (
            'MARKETING_EMAIL',
            'MARKETING_SMS',
            'ANALYTICS',
            'PERSONALIZATION',
            'THIRD_PARTY_SHARING',
            'ESSENTIAL'
        )
    )
);

CREATE INDEX IF NOT EXISTS idx_consent_history_customer_id ON consent_history (customer_id);
CREATE INDEX IF NOT EXISTS idx_consent_history_type ON consent_history (customer_id, type);
CREATE INDEX IF NOT EXISTS idx_consent_history_created_at ON consent_history (created_at DESC);
