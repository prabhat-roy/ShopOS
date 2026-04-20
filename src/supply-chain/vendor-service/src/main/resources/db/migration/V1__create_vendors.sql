-- V1__create_vendors.sql
-- Creates the vendors table for the vendor-service

CREATE TABLE IF NOT EXISTS vendors (
    id           UUID                        NOT NULL DEFAULT gen_random_uuid(),
    name         VARCHAR(255)                NOT NULL,
    email        VARCHAR(255)                NOT NULL,
    phone        VARCHAR(50),
    website      VARCHAR(255),
    status       VARCHAR(30)                 NOT NULL DEFAULT 'PENDING_APPROVAL',
    country      VARCHAR(100),
    address      TEXT,
    tax_id       VARCHAR(100),
    rating       NUMERIC(3, 2),
    total_orders INTEGER                     NOT NULL DEFAULT 0,
    created_at   TIMESTAMP WITH TIME ZONE    NOT NULL DEFAULT now(),
    updated_at   TIMESTAMP WITH TIME ZONE    NOT NULL DEFAULT now(),

    CONSTRAINT pk_vendors PRIMARY KEY (id),
    CONSTRAINT uq_vendors_email UNIQUE (email),
    CONSTRAINT chk_vendors_status CHECK (
        status IN ('ACTIVE', 'INACTIVE', 'SUSPENDED', 'PENDING_APPROVAL')
    ),
    CONSTRAINT chk_vendors_rating CHECK (
        rating IS NULL OR (rating >= 0.00 AND rating <= 5.00)
    ),
    CONSTRAINT chk_vendors_total_orders CHECK (total_orders >= 0)
);

-- Index on status for filtering by vendor state
CREATE INDEX IF NOT EXISTS idx_vendors_status ON vendors (status);

-- Index on country for geographic filtering
CREATE INDEX IF NOT EXISTS idx_vendors_country ON vendors (country);

-- Index on email (already covered by unique constraint, added explicitly for clarity)
CREATE INDEX IF NOT EXISTS idx_vendors_email ON vendors (email);

-- Index on created_at for sorting/pagination
CREATE INDEX IF NOT EXISTS idx_vendors_created_at ON vendors (created_at DESC);

COMMENT ON TABLE vendors IS 'Stores supply-chain vendor profiles managed by vendor-service';
COMMENT ON COLUMN vendors.id IS 'UUID primary key, auto-generated';
COMMENT ON COLUMN vendors.status IS 'Lifecycle state: PENDING_APPROVAL, ACTIVE, INACTIVE, SUSPENDED';
COMMENT ON COLUMN vendors.rating IS 'Aggregate rating 0.00–5.00, updated by review aggregation jobs';
COMMENT ON COLUMN vendors.total_orders IS 'Denormalised count of purchase orders placed with this vendor';
