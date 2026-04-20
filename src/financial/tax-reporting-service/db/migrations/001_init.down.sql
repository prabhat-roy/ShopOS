-- Migration: 001_init.down.sql
-- Reverses the 001_init.up.sql migration.

DROP INDEX IF EXISTS idx_tax_records_tax_type_date;
DROP INDEX IF EXISTS idx_tax_records_jurisdiction_date;
DROP INDEX IF EXISTS idx_tax_records_transaction_date;
DROP INDEX IF EXISTS idx_tax_records_tax_type;
DROP INDEX IF EXISTS idx_tax_records_jurisdiction;
DROP TABLE IF EXISTS tax_records;
