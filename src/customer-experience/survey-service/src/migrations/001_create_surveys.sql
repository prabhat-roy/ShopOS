-- Migration: 001_create_surveys
-- Description: Create surveys and survey_responses tables

CREATE TABLE IF NOT EXISTS surveys (
    id          UUID          NOT NULL DEFAULT gen_random_uuid(),
    title       VARCHAR(255)  NOT NULL,
    description TEXT,
    questions   JSONB         NOT NULL DEFAULT '[]',
    status      VARCHAR(20)   NOT NULL DEFAULT 'DRAFT',
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    CONSTRAINT surveys_pkey PRIMARY KEY (id),
    CONSTRAINT surveys_status_check CHECK (status IN ('DRAFT', 'ACTIVE', 'CLOSED')),
    CONSTRAINT surveys_title_length CHECK (length(title) >= 1)
);

CREATE INDEX IF NOT EXISTS idx_surveys_status ON surveys (status);
CREATE INDEX IF NOT EXISTS idx_surveys_created_at ON surveys (created_at DESC);

COMMENT ON COLUMN surveys.questions IS 'JSONB array of question objects: [{id, type, text, options?, required?}]';
COMMENT ON COLUMN surveys.status IS 'DRAFT → ACTIVE → CLOSED lifecycle';

CREATE TABLE IF NOT EXISTS survey_responses (
    id          UUID          NOT NULL DEFAULT gen_random_uuid(),
    survey_id   UUID          NOT NULL,
    customer_id VARCHAR(255),
    answers     JSONB         NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    CONSTRAINT survey_responses_pkey PRIMARY KEY (id),
    CONSTRAINT survey_responses_survey_fk FOREIGN KEY (survey_id)
        REFERENCES surveys (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_survey_responses_survey_id ON survey_responses (survey_id);
CREATE INDEX IF NOT EXISTS idx_survey_responses_customer_id ON survey_responses (customer_id);
CREATE INDEX IF NOT EXISTS idx_survey_responses_created_at ON survey_responses (created_at DESC);

COMMENT ON COLUMN survey_responses.answers IS 'JSONB map of questionId → answer value';
