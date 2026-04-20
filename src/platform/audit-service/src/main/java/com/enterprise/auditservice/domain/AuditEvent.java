package com.enterprise.auditservice.domain;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.Id;
import jakarta.persistence.Table;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.Instant;

/**
 * JPA entity representing a single immutable audit record.
 * Maps 1:1 to the audit_events table created by V1__create_audit_events.sql.
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(name = "audit_events")
public class AuditEvent {

    /** UUID assigned by the producing service or generated on ingest. */
    @Id
    @Column(name = "id", nullable = false)
    private String id;

    /** Identifier of the user or service that performed the action. */
    @Column(name = "actor_id", nullable = false)
    private String actorId;

    /** Discriminator: "user" | "service". */
    @Column(name = "actor_type", nullable = false)
    private String actorType;

    /** Dot-notated action name, e.g. "order.placed". */
    @Column(name = "action", nullable = false)
    private String action;

    /** Entity type that was acted upon, e.g. "Order", "User". */
    @Column(name = "resource_type", nullable = false)
    private String resourceType;

    /** Primary key of the acted-upon entity. */
    @Column(name = "resource_id", nullable = false)
    private String resourceId;

    /** Client IP address at the time of the action. */
    @Column(name = "ip_address", nullable = false)
    private String ipAddress;

    /** "success" or "failure". */
    @Column(name = "outcome", nullable = false)
    private String outcome;

    /**
     * Freeform JSONB payload forwarded verbatim from the Kafka message.
     * Stored as a JSON string; the columnDefinition tells Hibernate the
     * underlying column type so it does not try to cast it.
     */
    @Column(name = "metadata", nullable = false, columnDefinition = "jsonb")
    private String metadata;

    /** Source Kafka topic the event was consumed from. */
    @Column(name = "kafka_topic", nullable = false)
    private String kafkaTopic;

    /** Timestamp from the originating event (producer-set). */
    @Column(name = "occurred_at", nullable = false)
    private Instant occurredAt;

    /** Timestamp when the audit-service persisted the record (server-set). */
    @Column(name = "recorded_at", nullable = false)
    private Instant recordedAt;
}
