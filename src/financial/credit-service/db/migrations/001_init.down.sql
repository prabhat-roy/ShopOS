-- credit-service: rollback initial schema
-- Down migration: drops tables in dependency order.

DROP TABLE IF EXISTS credit_transactions;
DROP TABLE IF EXISTS credit_accounts;
