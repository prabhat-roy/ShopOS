-- 001_init.up.sql
-- Creates the translations table with unique constraint and indexes.

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS translations (
    id          UUID         NOT NULL DEFAULT gen_random_uuid(),
    locale      VARCHAR(10)  NOT NULL,
    namespace   VARCHAR(128) NOT NULL,
    key         VARCHAR(512) NOT NULL,
    value       TEXT         NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    PRIMARY KEY (id),
    CONSTRAINT uq_translation UNIQUE (locale, namespace, key)
);

CREATE INDEX IF NOT EXISTS idx_translations_locale     ON translations (locale);
CREATE INDEX IF NOT EXISTS idx_translations_namespace  ON translations (namespace);
CREATE INDEX IF NOT EXISTS idx_translations_locale_ns  ON translations (locale, namespace);

COMMENT ON TABLE  translations        IS 'Stores all i18n/l10n translation strings';
COMMENT ON COLUMN translations.locale    IS 'BCP-47 locale tag e.g. en, es, fr';
COMMENT ON COLUMN translations.namespace IS 'Logical grouping e.g. common, checkout, errors';
COMMENT ON COLUMN translations.key       IS 'Dot-notation translation key e.g. button.submit';
COMMENT ON COLUMN translations.value     IS 'Translated string for this locale/namespace/key';
