package com.enterprise.auditservice.service;

import com.enterprise.auditservice.domain.AuditEvent;
import com.enterprise.auditservice.dto.KafkaEventDto;
import com.enterprise.auditservice.repository.AuditEventRepository;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.apache.kafka.clients.consumer.ConsumerRecord;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.kafka.support.Acknowledgment;
import org.springframework.stereotype.Service;

import java.time.Instant;
import java.util.UUID;

/**
 * Kafka consumer that ingests auditable events from multiple topics across
 * all ShopOS domains and persists them as immutable {@link AuditEvent} rows.
 *
 * <p>Each listener method acknowledges the offset manually after a successful
 * database write, preventing message loss on restart while still processing
 * messages as fast as the thread pool allows.</p>
 *
 * <p>Malformed JSON payloads are logged and skipped rather than allowing a
 * poison-pill message to halt partition consumption.</p>
 */
@Slf4j
@Service
@RequiredArgsConstructor
public class AuditConsumerService {

    private static final String GROUP_ID = "audit-service";

    private final AuditEventRepository repository;
    private final ObjectMapper objectMapper;

    // ── Identity domain ────────────────────────────────────────────────────────

    @KafkaListener(
            topics = {"identity.user.registered", "identity.user.deleted"},
            groupId = GROUP_ID
    )
    public void consumeIdentityEvents(ConsumerRecord<String, String> record,
                                      Acknowledgment ack) {
        processRecord(record, ack);
    }

    // ── Commerce domain ────────────────────────────────────────────────────────

    @KafkaListener(
            topics = {
                    "commerce.order.placed",
                    "commerce.order.cancelled",
                    "commerce.payment.processed"
            },
            groupId = GROUP_ID
    )
    public void consumeCommerceEvents(ConsumerRecord<String, String> record,
                                      Acknowledgment ack) {
        processRecord(record, ack);
    }

    // ── Security domain ────────────────────────────────────────────────────────

    @KafkaListener(
            topics = {"security.fraud.detected", "security.login.failed"},
            groupId = GROUP_ID
    )
    public void consumeSecurityEvents(ConsumerRecord<String, String> record,
                                      Acknowledgment ack) {
        processRecord(record, ack);
    }

    // ── Core processing ────────────────────────────────────────────────────────

    /**
     * Deserializes the Kafka record value to {@link KafkaEventDto}, maps it
     * to an {@link AuditEvent}, persists it, then acknowledges the offset.
     *
     * <p>On {@link JsonProcessingException} the message is logged and skipped;
     * the offset is still acknowledged to avoid an infinite retry loop on a
     * dead-letter candidate.</p>
     *
     * @param record the raw Kafka consumer record
     * @param ack    manual acknowledgment handle
     */
    void processRecord(ConsumerRecord<String, String> record, Acknowledgment ack) {
        String topic = record.topic();
        String payload = record.value();

        log.debug("Received message on topic={} partition={} offset={}",
                topic, record.partition(), record.offset());

        try {
            KafkaEventDto dto = objectMapper.readValue(payload, KafkaEventDto.class);
            AuditEvent event = toEntity(dto, topic);
            repository.save(event);
            log.info("Persisted audit event id={} action={} topic={}",
                    event.getId(), event.getAction(), topic);
        } catch (JsonProcessingException ex) {
            log.error("Failed to deserialize Kafka message on topic={} offset={}: {}",
                    topic, record.offset(), ex.getMessage());
            // Acknowledge anyway — do not block the partition on a bad message.
        } finally {
            if (ack != null) {
                ack.acknowledge();
            }
        }
    }

    /**
     * Maps a {@link KafkaEventDto} to a new {@link AuditEvent} entity.
     * Null-safe defaults are applied so that the NOT NULL database constraints
     * are always satisfied regardless of how much the producer provides.
     *
     * @param dto   deserialized Kafka payload
     * @param topic source Kafka topic name
     * @return a fully populated, unsaved entity
     */
    private AuditEvent toEntity(KafkaEventDto dto, String topic) {
        return AuditEvent.builder()
                .id(UUID.randomUUID().toString())
                .actorId(nullToEmpty(dto.actorId()))
                .actorType(nullToEmpty(dto.actorType()))
                .action(dto.action() != null ? dto.action() : deriveActionFromTopic(topic))
                .resourceType(nullToEmpty(dto.resourceType()))
                .resourceId(nullToEmpty(dto.resourceId()))
                .ipAddress(nullToEmpty(dto.ipAddress()))
                .outcome(dto.outcome() != null ? dto.outcome() : "success")
                .metadata(dto.metadata() != null ? dto.metadata() : "{}")
                .kafkaTopic(topic)
                .occurredAt(dto.occurredAt() != null ? dto.occurredAt() : Instant.now())
                .recordedAt(Instant.now())
                .build();
    }

    /** Returns the topic suffix as a fallback action name (e.g. "order.placed"). */
    private static String deriveActionFromTopic(String topic) {
        int dot = topic.indexOf('.');
        return dot >= 0 ? topic.substring(dot + 1) : topic;
    }

    private static String nullToEmpty(String value) {
        return value != null ? value : "";
    }
}
