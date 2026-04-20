-- ============================================================
-- V1__create_kyc_aml.sql
-- KYC records and AML checks schema for kyc-aml-service
-- ============================================================

-- KYC Records table
CREATE TABLE kyc_records (
    id               UUID         NOT NULL DEFAULT gen_random_uuid(),
    customer_id      UUID         NOT NULL,
    first_name       VARCHAR(100) NOT NULL,
    last_name        VARCHAR(100) NOT NULL,
    date_of_birth    DATE         NOT NULL,
    nationality      CHAR(2)      NOT NULL,
    document_type    VARCHAR(30)  NOT NULL,
    document_number  VARCHAR(100) NOT NULL,
    document_expiry  DATE         NOT NULL,
    status           VARCHAR(20)  NOT NULL DEFAULT 'PENDING',
    risk_level       VARCHAR(10),
    verified_at      TIMESTAMP,
    expires_at       TIMESTAMP,
    rejection_reason TEXT,
    notes            TEXT,
    created_at       TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMP    NOT NULL DEFAULT NOW(),

    CONSTRAINT pk_kyc_records PRIMARY KEY (id),
    CONSTRAINT uq_kyc_customer_id UNIQUE (customer_id),
    CONSTRAINT chk_kyc_status CHECK (
        status IN ('PENDING','IN_PROGRESS','VERIFIED','REJECTED','EXPIRED','SUSPENDED')
    ),
    CONSTRAINT chk_kyc_risk_level CHECK (
        risk_level IS NULL OR risk_level IN ('LOW','MEDIUM','HIGH','CRITICAL')
    ),
    CONSTRAINT chk_kyc_document_type CHECK (
        document_type IN ('PASSPORT','NATIONAL_ID','DRIVERS_LICENSE')
    )
);

CREATE INDEX idx_kyc_records_customer_id ON kyc_records (customer_id);
CREATE INDEX idx_kyc_records_status      ON kyc_records (status);
CREATE INDEX idx_kyc_records_expires_at  ON kyc_records (expires_at)
    WHERE status = 'VERIFIED';

-- AML Checks table
CREATE TABLE aml_checks (
    id          UUID         NOT NULL DEFAULT gen_random_uuid(),
    customer_id UUID         NOT NULL,
    check_type  VARCHAR(40)  NOT NULL,
    result      VARCHAR(20)  NOT NULL,
    risk_score  SMALLINT     NOT NULL CHECK (risk_score BETWEEN 0 AND 100),
    details     TEXT,
    checked_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMP,
    resolved_by VARCHAR(200),
    resolution  TEXT,

    CONSTRAINT pk_aml_checks PRIMARY KEY (id),
    CONSTRAINT chk_aml_check_type CHECK (
        check_type IN ('SANCTIONS','PEP','ADVERSE_MEDIA','TRANSACTION_MONITORING')
    ),
    CONSTRAINT chk_aml_result CHECK (
        result IN ('CLEAR','FLAGGED','REVIEW_REQUIRED')
    )
);

CREATE INDEX idx_aml_checks_customer_id       ON aml_checks (customer_id);
CREATE INDEX idx_aml_checks_check_type_result ON aml_checks (check_type, result);
CREATE INDEX idx_aml_checks_checked_at        ON aml_checks (checked_at DESC);
