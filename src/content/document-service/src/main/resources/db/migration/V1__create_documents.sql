-- V1__create_documents.sql
-- Creates the documents metadata table

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS documents (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id        UUID         NOT NULL,
    name            VARCHAR(512) NOT NULL,
    original_name   VARCHAR(512) NOT NULL,
    content_type    VARCHAR(128) NOT NULL,
    size            BIGINT       NOT NULL CHECK (size >= 0),
    document_type   VARCHAR(20)  NOT NULL
                        CHECK (document_type IN ('PDF','WORD','EXCEL','TEXT','IMAGE','OTHER')),
    stored_key      TEXT         NOT NULL UNIQUE,
    tags            TEXT,
    description     TEXT,
    version         INT          NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_documents_owner_id    ON documents (owner_id);
CREATE INDEX idx_documents_type        ON documents (document_type);
CREATE INDEX idx_documents_created_at  ON documents (created_at DESC);
CREATE INDEX idx_documents_name_lower  ON documents (LOWER(name));

COMMENT ON TABLE documents IS 'Document metadata — file binaries live in MinIO';
COMMENT ON COLUMN documents.stored_key IS 'MinIO object key: documents/{ownerId}/{uuid}/{filename}';
