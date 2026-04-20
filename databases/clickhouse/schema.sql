-- ClickHouse schema for ShopOS OLAP / reporting

CREATE DATABASE IF NOT EXISTS shopos;

-- ── Orders fact table ─────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS shopos.orders (
    order_id        UUID,
    user_id         UUID,
    tenant_id       String,
    status          LowCardinality(String),
    total_amount    Decimal(18, 2),
    currency        LowCardinality(String),
    item_count      UInt16,
    created_at      DateTime,
    fulfilled_at    Nullable(DateTime),
    cancelled_at    Nullable(DateTime),
    date            Date MATERIALIZED toDate(created_at)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(created_at)
ORDER BY (tenant_id, created_at, order_id)
TTL created_at + INTERVAL 3 YEAR;

-- ── Events fact table ─────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS shopos.events (
    event_id        UUID,
    event_type      LowCardinality(String),
    user_id         Nullable(UUID),
    session_id      Nullable(String),
    tenant_id       String,
    properties      String,  -- JSON
    created_at      DateTime,
    date            Date MATERIALIZED toDate(created_at)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(created_at)
ORDER BY (tenant_id, event_type, created_at)
TTL created_at + INTERVAL 1 YEAR;

-- ── Revenue aggregation (materialized view) ───────────────────────────────────
CREATE TABLE IF NOT EXISTS shopos.revenue_daily (
    tenant_id       String,
    date            Date,
    currency        LowCardinality(String),
    order_count     UInt64,
    total_revenue   Decimal(18, 2),
    avg_order_value Decimal(18, 2)
) ENGINE = SummingMergeTree((order_count, total_revenue))
ORDER BY (tenant_id, date, currency);

CREATE MATERIALIZED VIEW IF NOT EXISTS shopos.revenue_daily_mv
TO shopos.revenue_daily AS
SELECT
    tenant_id,
    toDate(created_at)           AS date,
    currency,
    count()                      AS order_count,
    sum(total_amount)            AS total_revenue,
    avg(total_amount)            AS avg_order_value
FROM shopos.orders
WHERE status = 'fulfilled'
GROUP BY tenant_id, date, currency;

-- ── Product click analytics ───────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS shopos.product_clicks (
    product_id      UUID,
    user_id         Nullable(UUID),
    tenant_id       String,
    source          LowCardinality(String),
    created_at      DateTime,
    date            Date MATERIALIZED toDate(created_at)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(created_at)
ORDER BY (tenant_id, product_id, created_at)
TTL created_at + INTERVAL 6 MONTH;
