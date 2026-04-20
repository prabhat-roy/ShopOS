-- ─────────────────────────────────────────────────────────────────────────────
-- V1__create_audit_events.sql
-- Creates the core audit_events table and supporting indexes.
-- All significant actions across every ShopOS domain land here for compliance
-- and security review.
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE audit_events (
    id            TEXT        PRIMARY KEY,
    actor_id      TEXT        NOT NULL,
    actor_type    TEXT        NOT NULL,
    action        TEXT        NOT NULL,
    resource_type TEXT        NOT NULL DEFAULT '',
    resource_id   TEXT        NOT NULL DEFAULT '',
    ip_address    TEXT        NOT NULL DEFAULT '',
    outcome       TEXT        NOT NULL DEFAULT 'success',
    metadata      JSONB       NOT NULL DEFAULT '{}',
    kafka_topic   TEXT        NOT NULL DEFAULT '',
    occurred_at   TIMESTAMPTZ NOT NULL,
    recorded_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ── Indexes ───────────────────────────────────────────────────────────────────
-- Look up all events triggered by a specific actor (user or service)
CREATE INDEX idx_audit_actor_id    ON audit_events (actor_id);

-- Look up all events targeting a specific resource (e.g., a specific Order UUID)
CREATE INDEX idx_audit_resource    ON audit_events (resource_type, resource_id);

-- Time-range queries for compliance reports
CREATE INDEX idx_audit_occurred    ON audit_events (occurred_at);

-- Filter by action type (e.g., "security.login.failed")
CREATE INDEX idx_audit_action      ON audit_events (action);

-- Composite index for actor + time range — common compliance query pattern
CREATE INDEX idx_audit_actor_time  ON audit_events (actor_id, occurred_at);

-- ── Comments ──────────────────────────────────────────────────────────────────
COMMENT ON TABLE  audit_events                IS 'Immutable log of every auditable action across all ShopOS domains.';
COMMENT ON COLUMN audit_events.id             IS 'UUID assigned by the producing service or generated on ingest.';
COMMENT ON COLUMN audit_events.actor_id       IS 'Identifier of the user or service that performed the action.';
COMMENT ON COLUMN audit_events.actor_type     IS 'Discriminator: "user" | "service".';
COMMENT ON COLUMN audit_events.action         IS 'Dot-notated action name, mirrors Kafka topic suffix, e.g. "order.placed".';
COMMENT ON COLUMN audit_events.resource_type  IS 'Entity type that was acted upon, e.g. "Order", "User".';
COMMENT ON COLUMN audit_events.resource_id    IS 'Primary key of the acted-upon entity.';
COMMENT ON COLUMN audit_events.ip_address     IS 'Client IP address at the time of the action (empty for internal service calls).';
COMMENT ON COLUMN audit_events.outcome        IS '"success" or "failure".';
COMMENT ON COLUMN audit_events.metadata       IS 'Freeform JSONB payload forwarded verbatim from the Kafka message.';
COMMENT ON COLUMN audit_events.kafka_topic    IS 'Source Kafka topic the event was consumed from.';
COMMENT ON COLUMN audit_events.occurred_at    IS 'Timestamp from the originating event (producer-set).';
COMMENT ON COLUMN audit_events.recorded_at    IS 'Timestamp when the audit-service persisted the record (server-set).';
