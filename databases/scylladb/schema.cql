-- ScyllaDB schema for ShopOS high-performance time-series and analytics
-- ScyllaDB is Cassandra-compatible; replace Cassandra keyspaces with these for services
-- that need higher throughput on the same data model.

-- ── Keyspace ──────────────────────────────────────────────────────────────────
CREATE KEYSPACE IF NOT EXISTS shopos
  WITH replication = {
    'class': 'NetworkTopologyStrategy',
    'datacenter1': 3
  }
  AND durable_writes = true;

USE shopos;

-- ── Analytics events (replaces Cassandra analytics_events) ───────────────────
CREATE TABLE IF NOT EXISTS events (
    tenant_id   text,
    event_type  text,
    date        date,
    created_at  timestamp,
    event_id    uuid,
    user_id     uuid,
    session_id  text,
    properties  text,   -- JSON blob
    PRIMARY KEY ((tenant_id, event_type, date), created_at, event_id)
) WITH CLUSTERING ORDER BY (created_at DESC, event_id ASC)
  AND default_time_to_live = 7776000  -- 90 days
  AND compaction = {
    'class': 'TimeWindowCompactionStrategy',
    'compaction_window_unit': 'DAYS',
    'compaction_window_size': 1
  };

-- ── Page views ────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS page_views (
    tenant_id   text,
    date        date,
    created_at  timestamp,
    session_id  text,
    user_id     uuid,
    page_url    text,
    referrer    text,
    duration_ms int,
    PRIMARY KEY ((tenant_id, date), created_at, session_id)
) WITH CLUSTERING ORDER BY (created_at DESC)
  AND default_time_to_live = 2592000  -- 30 days
  AND compaction = {
    'class': 'TimeWindowCompactionStrategy',
    'compaction_window_unit': 'DAYS',
    'compaction_window_size': 1
  };

-- ── Product click stream ──────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS product_clicks (
    tenant_id   text,
    product_id  uuid,
    date        date,
    created_at  timestamp,
    click_id    uuid,
    user_id     uuid,
    source      text,
    PRIMARY KEY ((tenant_id, product_id, date), created_at, click_id)
) WITH CLUSTERING ORDER BY (created_at DESC)
  AND default_time_to_live = 15552000  -- 180 days
  AND compaction = {
    'class': 'TimeWindowCompactionStrategy',
    'compaction_window_unit': 'DAYS',
    'compaction_window_size': 7
  };

-- ── Search queries ────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS search_queries (
    tenant_id    text,
    date         date,
    created_at   timestamp,
    query_id     uuid,
    user_id      uuid,
    query_text   text,
    result_count int,
    clicked      boolean,
    PRIMARY KEY ((tenant_id, date), created_at, query_id)
) WITH CLUSTERING ORDER BY (created_at DESC)
  AND default_time_to_live = 7776000;  -- 90 days

-- ── Real-time counters (materialised aggregates) ──────────────────────────────
CREATE TABLE IF NOT EXISTS product_view_counts (
    tenant_id  text,
    product_id uuid,
    date       date,
    view_count counter,
    PRIMARY KEY ((tenant_id, product_id), date)
);

CREATE TABLE IF NOT EXISTS search_query_counts (
    tenant_id  text,
    date       date,
    query_text text,
    hit_count  counter,
    PRIMARY KEY ((tenant_id, date), query_text)
);
