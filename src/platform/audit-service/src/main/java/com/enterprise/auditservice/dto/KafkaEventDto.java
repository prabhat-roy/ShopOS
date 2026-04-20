package com.enterprise.auditservice.dto;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

import java.time.Instant;

/**
 * Immutable DTO for deserializing incoming Kafka JSON payloads.
 *
 * <p>Fields are nullable to be tolerant of partial payloads from different
 * producer services; the consumer service normalises missing values before
 * persisting.</p>
 *
 * <p>{@code @JsonIgnoreProperties(ignoreUnknown = true)} ensures that extra
 * fields present in some producers do not break deserialization.</p>
 */
@JsonIgnoreProperties(ignoreUnknown = true)
public record KafkaEventDto(
        String actorId,
        String actorType,
        String action,
        String resourceType,
        String resourceId,
        String ipAddress,
        String outcome,
        String metadata,
        Instant occurredAt
) {}
