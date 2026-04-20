-- Migration: 001_init.down.sql
-- Reverses the 001_init.up.sql migration.

DROP INDEX IF EXISTS idx_expenses_created_at;
DROP INDEX IF EXISTS idx_expenses_employee_status;
DROP INDEX IF EXISTS idx_expenses_category;
DROP INDEX IF EXISTS idx_expenses_status;
DROP INDEX IF EXISTS idx_expenses_employee_id;
DROP TABLE IF EXISTS expenses;
