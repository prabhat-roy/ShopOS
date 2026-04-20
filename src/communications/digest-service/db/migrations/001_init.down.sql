-- Drop indexes first, then tables in dependency order.
DROP INDEX IF EXISTS idx_digest_runs_config_id_sent_at;
DROP INDEX IF EXISTS idx_digest_configs_user_id;
DROP INDEX IF EXISTS idx_digest_configs_due;

DROP TABLE IF EXISTS digest_runs;
DROP TABLE IF EXISTS digest_configs;
