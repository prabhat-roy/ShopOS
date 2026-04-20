-- 001_init.down.sql
-- Reverses the 001_init migration.

DROP INDEX IF EXISTS idx_translations_locale_ns;
DROP INDEX IF EXISTS idx_translations_namespace;
DROP INDEX IF EXISTS idx_translations_locale;
DROP TABLE IF EXISTS translations;
