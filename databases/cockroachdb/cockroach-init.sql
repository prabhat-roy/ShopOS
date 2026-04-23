-- CockroachDB initialization SQL for ShopOS
-- Run after cluster init: cockroach sql --insecure --host=cockroachdb:26257

-- Create the main shopos database
CREATE DATABASE IF NOT EXISTS shopos;

-- Create application user
CREATE USER IF NOT EXISTS shopos_app WITH PASSWORD 'changeme-replace-in-production';

-- Grant privileges
GRANT ALL ON DATABASE shopos TO shopos_app;

-- Create schemas for each domain
\c shopos;

CREATE SCHEMA IF NOT EXISTS commerce;
CREATE SCHEMA IF NOT EXISTS catalog;
CREATE SCHEMA IF NOT EXISTS identity;
CREATE SCHEMA IF NOT EXISTS financial;
CREATE SCHEMA IF NOT EXISTS supply_chain;

-- Grant schema permissions
GRANT ALL ON SCHEMA commerce TO shopos_app;
GRANT ALL ON SCHEMA catalog TO shopos_app;
GRANT ALL ON SCHEMA identity TO shopos_app;
GRANT ALL ON SCHEMA financial TO shopos_app;
GRANT ALL ON SCHEMA supply_chain TO shopos_app;

-- Enable follower reads for analytics
SET CLUSTER SETTING kv.rangefeed.enabled = true;

-- Set replication factor
ALTER DATABASE shopos CONFIGURE ZONE USING num_replicas = 3;
