-- budget-service: rollback initial schema
-- Down migration: drops tables in dependency order.

DROP TABLE IF EXISTS spending_records;
DROP TABLE IF EXISTS budget_allocations;
DROP TABLE IF EXISTS budgets;
