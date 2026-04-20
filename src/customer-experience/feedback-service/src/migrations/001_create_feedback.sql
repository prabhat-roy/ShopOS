-- Migration: 001_create_feedback
-- Description: Create feedback table for product and service feedback

CREATE TABLE IF NOT EXISTS feedback (
    id            UUID          NOT NULL DEFAULT gen_random_uuid(),
    customer_id   VARCHAR(255),
    type          VARCHAR(30)   NOT NULL,
    status        VARCHAR(20)   NOT NULL DEFAULT 'NEW',
    score         SMALLINT,
    title         VARCHAR(500),
    body          TEXT,
    contact_email VARCHAR(320),
    metadata      JSONB,
    note          TEXT,
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    CONSTRAINT feedback_pkey PRIMARY KEY (id),
    CONSTRAINT feedback_type_check CHECK (
        type IN ('NPS', 'FEATURE_REQUEST', 'BUG_REPORT', 'GENERAL', 'COMPLAINT')
    ),
    CONSTRAINT feedback_status_check CHECK (
        status IN ('NEW', 'REVIEWED', 'IN_PROGRESS', 'RESOLVED', 'CLOSED')
    ),
    CONSTRAINT feedback_nps_score_check CHECK (
        (type != 'NPS') OR (score IS NOT NULL AND score >= 0 AND score <= 10)
    )
);

CREATE INDEX IF NOT EXISTS idx_feedback_type ON feedback (type);
CREATE INDEX IF NOT EXISTS idx_feedback_status ON feedback (status);
CREATE INDEX IF NOT EXISTS idx_feedback_customer_id ON feedback (customer_id);
CREATE INDEX IF NOT EXISTS idx_feedback_created_at ON feedback (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_feedback_type_status ON feedback (type, status);

COMMENT ON COLUMN feedback.score IS 'NPS score 0-10: 0-6 detractor, 7-8 passive, 9-10 promoter';
COMMENT ON COLUMN feedback.metadata IS 'Arbitrary structured data (product_id, order_id, page, etc.)';
COMMENT ON COLUMN feedback.note IS 'Internal resolution note added by support agent';
