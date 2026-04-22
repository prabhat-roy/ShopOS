-- TimescaleDB schema for ShopOS time-series data
-- Used by: analytics-service, event-tracking-service, metrics aggregation
-- Requires TimescaleDB extension on PostgreSQL 16

CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;

-- ─── Order metrics time-series ────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS order_metrics (
    time            TIMESTAMPTZ NOT NULL,
    domain          TEXT        NOT NULL,   -- 'commerce', 'marketplace', etc.
    metric          TEXT        NOT NULL,   -- 'orders_placed', 'revenue', 'avg_order_value'
    value           DOUBLE PRECISION NOT NULL,
    currency        TEXT        DEFAULT 'USD',
    region          TEXT,
    tags            JSONB       DEFAULT '{}'
);

SELECT create_hypertable('order_metrics', 'time', if_not_exists => TRUE);

CREATE INDEX idx_order_metrics_domain ON order_metrics (domain, time DESC);
CREATE INDEX idx_order_metrics_metric ON order_metrics (metric, time DESC);

-- Continuous aggregate: hourly revenue
CREATE MATERIALIZED VIEW order_metrics_hourly
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', time) AS bucket,
    domain,
    metric,
    SUM(value)   AS total,
    AVG(value)   AS avg,
    COUNT(*)     AS count,
    MAX(value)   AS max,
    MIN(value)   AS min
FROM order_metrics
GROUP BY bucket, domain, metric
WITH NO DATA;

SELECT add_continuous_aggregate_policy('order_metrics_hourly',
    start_offset => INTERVAL '1 week',
    end_offset   => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour');

-- ─── Page view events ─────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS page_views (
    time            TIMESTAMPTZ NOT NULL,
    session_id      UUID,
    user_id         UUID,
    path            TEXT        NOT NULL,
    referrer        TEXT,
    device_type     TEXT,       -- 'mobile', 'tablet', 'desktop'
    country         TEXT,
    duration_ms     INTEGER,
    bounce          BOOLEAN     DEFAULT FALSE
);

SELECT create_hypertable('page_views', 'time', if_not_exists => TRUE);

CREATE INDEX idx_page_views_path ON page_views (path, time DESC);
CREATE INDEX idx_page_views_user ON page_views (user_id, time DESC);

-- ─── Service performance metrics ──────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS service_metrics (
    time            TIMESTAMPTZ NOT NULL,
    service         TEXT        NOT NULL,
    endpoint        TEXT        NOT NULL,
    method          TEXT        NOT NULL,   -- 'GET', 'POST', etc.
    status_code     INTEGER     NOT NULL,
    duration_ms     DOUBLE PRECISION NOT NULL,
    request_size    BIGINT,
    response_size   BIGINT,
    trace_id        TEXT,
    pod             TEXT
);

SELECT create_hypertable('service_metrics', 'time', if_not_exists => TRUE);

CREATE INDEX idx_svc_metrics_service ON service_metrics (service, time DESC);
CREATE INDEX idx_svc_metrics_status  ON service_metrics (status_code, time DESC);

-- Continuous aggregate: per-service p99 latency per minute
CREATE MATERIALIZED VIEW service_latency_1m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 minute', time) AS bucket,
    service,
    endpoint,
    COUNT(*)                          AS request_count,
    AVG(duration_ms)                  AS avg_latency,
    percentile_cont(0.50) WITHIN GROUP (ORDER BY duration_ms) AS p50,
    percentile_cont(0.95) WITHIN GROUP (ORDER BY duration_ms) AS p95,
    percentile_cont(0.99) WITHIN GROUP (ORDER BY duration_ms) AS p99,
    SUM(CASE WHEN status_code >= 500 THEN 1 ELSE 0 END) AS error_count
FROM service_metrics
GROUP BY bucket, service, endpoint
WITH NO DATA;

SELECT add_continuous_aggregate_policy('service_latency_1m',
    start_offset => INTERVAL '2 days',
    end_offset   => INTERVAL '1 minute',
    schedule_interval => INTERVAL '1 minute');

-- ─── Inventory time-series ────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS inventory_events (
    time            TIMESTAMPTZ NOT NULL,
    product_id      UUID        NOT NULL,
    sku             TEXT        NOT NULL,
    warehouse_id    TEXT        NOT NULL,
    quantity_delta  INTEGER     NOT NULL,   -- positive=restock, negative=sale/reserve
    event_type      TEXT        NOT NULL,   -- 'sale', 'restock', 'reserve', 'release', 'adjustment'
    quantity_after  INTEGER     NOT NULL,
    reference_id    TEXT        -- order_id, po_id, etc.
);

SELECT create_hypertable('inventory_events', 'time', if_not_exists => TRUE);

CREATE INDEX idx_inventory_product ON inventory_events (product_id, time DESC);
CREATE INDEX idx_inventory_sku     ON inventory_events (sku, time DESC);

-- ─── Retention policies ───────────────────────────────────────────────────────
SELECT add_retention_policy('page_views',      INTERVAL '90 days');
SELECT add_retention_policy('service_metrics', INTERVAL '30 days');
SELECT add_retention_policy('order_metrics',   INTERVAL '2 years');
SELECT add_retention_policy('inventory_events', INTERVAL '1 year');

-- ─── Compression policies ─────────────────────────────────────────────────────
ALTER TABLE service_metrics SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'service, endpoint',
    timescaledb.compress_orderby = 'time DESC'
);

SELECT add_compression_policy('service_metrics', INTERVAL '7 days');

ALTER TABLE page_views SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'path, device_type',
    timescaledb.compress_orderby = 'time DESC'
);

SELECT add_compression_policy('page_views', INTERVAL '7 days');
