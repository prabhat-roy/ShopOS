-- Migration: 001_init.down.sql
-- Rolls back the initial schema for the returns-logistics-service.

BEGIN;

DROP INDEX IF EXISTS idx_return_authorizations_items;
DROP INDEX IF EXISTS idx_return_authorizations_created_at;
DROP INDEX IF EXISTS idx_return_authorizations_status;
DROP INDEX IF EXISTS idx_return_authorizations_order_id;
DROP INDEX IF EXISTS idx_return_authorizations_customer_id;

DROP TABLE IF EXISTS return_authorizations;

COMMIT;
