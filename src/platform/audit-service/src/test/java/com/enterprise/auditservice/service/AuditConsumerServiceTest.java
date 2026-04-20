package com.enterprise.auditservice.service;

import com.enterprise.auditservice.domain.AuditEvent;
import com.enterprise.auditservice.repository.AuditEventRepository;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import org.apache.kafka.clients.consumer.ConsumerRecord;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.kafka.support.Acknowledgment;

import java.time.Instant;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatNoException;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;

/**
 * Unit tests for {@link AuditConsumerService}.
 *
 * <p>The repository and Kafka acknowledgment are mocked; the real
 * {@link ObjectMapper} (configured with the JSR-310 module) is used to keep
 * JSON parsing behaviour realistic.</p>
 */
@ExtendWith(MockitoExtension.class)
class AuditConsumerServiceTest {

    @Mock
    private AuditEventRepository repository;

    @Mock
    private Acknowledgment ack;

    private AuditConsumerService service;

    private static final String TOPIC = "commerce.order.placed";

    @BeforeEach
    void setUp() {
        ObjectMapper mapper = new ObjectMapper()
                .registerModule(new JavaTimeModule())
                .disable(SerializationFeature.WRITE_DATES_AS_TIMESTAMPS);
        service = new AuditConsumerService(repository, mapper);
    }

    // ── Happy path ─────────────────────────────────────────────────────────────

    @Test
    @DisplayName("Valid JSON payload is persisted with the correct actorId")
    void validPayload_persistsCorrectActorId() {
        String payload = """
                {
                  "actorId":     "user-123",
                  "actorType":   "user",
                  "action":      "order.placed",
                  "resourceType":"Order",
                  "resourceId":  "order-456",
                  "ipAddress":   "10.0.0.1",
                  "outcome":     "success",
                  "metadata":    "{\\"items\\":3}",
                  "occurredAt":  "2024-06-01T12:00:00Z"
                }
                """;

        ConsumerRecord<String, String> record = new ConsumerRecord<>(TOPIC, 0, 0L, null, payload);

        service.processRecord(record, ack);

        ArgumentCaptor<AuditEvent> captor = ArgumentCaptor.forClass(AuditEvent.class);
        verify(repository, times(1)).save(captor.capture());

        AuditEvent saved = captor.getValue();
        assertThat(saved.getActorId()).isEqualTo("user-123");
        assertThat(saved.getActorType()).isEqualTo("user");
        assertThat(saved.getAction()).isEqualTo("order.placed");
        assertThat(saved.getResourceType()).isEqualTo("Order");
        assertThat(saved.getResourceId()).isEqualTo("order-456");
        assertThat(saved.getOutcome()).isEqualTo("success");
        assertThat(saved.getKafkaTopic()).isEqualTo(TOPIC);
        assertThat(saved.getId()).isNotBlank();
        assertThat(saved.getRecordedAt()).isNotNull();
    }

    @Test
    @DisplayName("Valid payload triggers acknowledgment after save")
    void validPayload_acknowledgesOffset() {
        String payload = """
                {
                  "actorId":   "svc-fraud",
                  "actorType": "service",
                  "action":    "fraud.detected",
                  "occurredAt":"2024-06-01T08:00:00Z"
                }
                """;

        ConsumerRecord<String, String> record =
                new ConsumerRecord<>("security.fraud.detected", 0, 1L, null, payload);

        service.processRecord(record, ack);

        verify(ack, times(1)).acknowledge();
    }

    @Test
    @DisplayName("Null fields in payload default to empty strings / 'success'")
    void partialPayload_defaultsApplied() {
        // Only mandatory fields present — all optional fields omitted
        String payload = """
                {
                  "actorId":   "user-789",
                  "actorType": "user",
                  "occurredAt":"2024-07-15T09:30:00Z"
                }
                """;

        ConsumerRecord<String, String> record =
                new ConsumerRecord<>(TOPIC, 0, 2L, null, payload);

        service.processRecord(record, ack);

        ArgumentCaptor<AuditEvent> captor = ArgumentCaptor.forClass(AuditEvent.class);
        verify(repository).save(captor.capture());

        AuditEvent saved = captor.getValue();
        assertThat(saved.getResourceType()).isEqualTo("");
        assertThat(saved.getResourceId()).isEqualTo("");
        assertThat(saved.getIpAddress()).isEqualTo("");
        assertThat(saved.getOutcome()).isEqualTo("success");
        assertThat(saved.getMetadata()).isEqualTo("{}");
        // action falls back to topic suffix when not provided
        assertThat(saved.getAction()).isEqualTo("order.placed");
    }

    // ── Error handling ─────────────────────────────────────────────────────────

    @Test
    @DisplayName("Invalid JSON does not throw and skips repository.save()")
    void invalidJson_noExceptionThrown_saveNotCalled() {
        String badPayload = "{ this is not valid json }";

        ConsumerRecord<String, String> record =
                new ConsumerRecord<>(TOPIC, 0, 3L, null, badPayload);

        assertThatNoException().isThrownBy(() -> service.processRecord(record, ack));
        verify(repository, never()).save(any());
    }

    @Test
    @DisplayName("Invalid JSON still acknowledges the offset to avoid blocking the partition")
    void invalidJson_offsetStillAcknowledged() {
        ConsumerRecord<String, String> record =
                new ConsumerRecord<>(TOPIC, 0, 4L, null, "not-json");

        service.processRecord(record, ack);

        verify(ack, times(1)).acknowledge();
    }

    @Test
    @DisplayName("Null payload is handled without throwing")
    void nullPayload_noExceptionThrown() {
        ConsumerRecord<String, String> record =
                new ConsumerRecord<>(TOPIC, 0, 5L, null, null);

        assertThatNoException().isThrownBy(() -> service.processRecord(record, ack));
        verify(repository, never()).save(any());
    }

    // ── Topic routing ──────────────────────────────────────────────────────────

    @Test
    @DisplayName("kafkaTopic field on saved entity matches the source topic")
    void kafkaTopic_setCorrectlyOnEntity() {
        String topic = "identity.user.registered";
        String payload = """
                {
                  "actorId":   "user-001",
                  "actorType": "user",
                  "action":    "user.registered",
                  "occurredAt":"2024-01-01T00:00:00Z"
                }
                """;

        ConsumerRecord<String, String> record = new ConsumerRecord<>(topic, 0, 6L, null, payload);
        service.processRecord(record, ack);

        ArgumentCaptor<AuditEvent> captor = ArgumentCaptor.forClass(AuditEvent.class);
        verify(repository).save(captor.capture());
        assertThat(captor.getValue().getKafkaTopic()).isEqualTo(topic);
    }

    @Test
    @DisplayName("recordedAt is set to current time and is not null")
    void recordedAt_isSetToNow() {
        Instant before = Instant.now();

        String payload = """
                {
                  "actorId":   "user-002",
                  "actorType": "user",
                  "occurredAt":"2024-03-10T10:00:00Z"
                }
                """;

        ConsumerRecord<String, String> record = new ConsumerRecord<>(TOPIC, 0, 7L, null, payload);
        service.processRecord(record, ack);

        Instant after = Instant.now();

        ArgumentCaptor<AuditEvent> captor = ArgumentCaptor.forClass(AuditEvent.class);
        verify(repository).save(captor.capture());

        Instant recordedAt = captor.getValue().getRecordedAt();
        assertThat(recordedAt).isNotNull();
        assertThat(recordedAt).isAfterOrEqualTo(before);
        assertThat(recordedAt).isBeforeOrEqualTo(after);
    }
}
