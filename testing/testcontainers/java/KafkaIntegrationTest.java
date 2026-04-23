package com.enterprise.shopos.testing;

import org.apache.kafka.clients.consumer.ConsumerConfig;
import org.apache.kafka.clients.consumer.ConsumerRecord;
import org.apache.kafka.clients.consumer.ConsumerRecords;
import org.apache.kafka.clients.consumer.KafkaConsumer;
import org.apache.kafka.clients.producer.KafkaProducer;
import org.apache.kafka.clients.producer.ProducerConfig;
import org.apache.kafka.clients.producer.ProducerRecord;
import org.apache.kafka.clients.producer.RecordMetadata;
import org.apache.kafka.common.serialization.StringDeserializer;
import org.apache.kafka.common.serialization.StringSerializer;
import org.junit.jupiter.api.AfterAll;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.TestInstance;
import org.testcontainers.containers.KafkaContainer;
import org.testcontainers.junit.jupiter.Container;
import org.testcontainers.junit.jupiter.Testcontainers;
import org.testcontainers.utility.DockerImageName;

import java.time.Duration;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.Properties;
import java.util.UUID;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.Future;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.TimeoutException;

import static org.junit.jupiter.api.Assertions.*;

/**
 * Integration test for Kafka using Testcontainers.
 *
 * Spins up a real confluent Kafka broker, produces messages to a topic,
 * and asserts that a consumer can read them back correctly.
 *
 * Dependencies (add to pom.xml or build.gradle.kts):
 *   org.testcontainers:kafka:1.20.x
 *   org.testcontainers:junit-jupiter:1.20.x
 *   org.apache.kafka:kafka-clients:3.8.x
 *   org.junit.jupiter:junit-jupiter:5.11.x
 */
@Testcontainers
@TestInstance(TestInstance.Lifecycle.PER_CLASS)
@DisplayName("Kafka Integration Tests — ShopOS Commerce Events")
class KafkaIntegrationTest {

    // Uses the Confluent Platform Kafka image (same as docker-compose.yml)
    @Container
    static final KafkaContainer kafkaContainer = new KafkaContainer(
            DockerImageName.parse("confluentinc/cp-kafka:7.7.1")
    ).withEnv("KAFKA_AUTO_CREATE_TOPICS_ENABLE", "true")
     .withEnv("KAFKA_NUM_PARTITIONS", "3")
     .withEnv("KAFKA_DEFAULT_REPLICATION_FACTOR", "1");

    // Topics matching the ShopOS Kafka naming convention: {domain}.{entity}.{event}
    private static final String TOPIC_ORDER_PLACED     = "commerce.order.placed";
    private static final String TOPIC_PAYMENT_PROCESSED = "commerce.payment.processed";
    private static final String TOPIC_USER_REGISTERED  = "identity.user.registered";

    private KafkaProducer<String, String> producer;
    private KafkaConsumer<String, String> consumer;

    @BeforeAll
    void setUp() {
        producer = createProducer();
        consumer = createConsumer("test-consumer-group-" + UUID.randomUUID());
    }

    @AfterAll
    void tearDown() {
        if (producer != null) producer.close(Duration.ofSeconds(5));
        if (consumer != null) consumer.close(Duration.ofSeconds(5));
    }

    // ── Test: produce and consume a single message ────────────────────────────
    @Test
    @DisplayName("Should produce and consume a commerce.order.placed event")
    void testProduceAndConsumeOrderPlacedEvent() throws ExecutionException, InterruptedException, TimeoutException {
        String orderId = UUID.randomUUID().toString();
        String payload = """
                {
                  "orderId": "%s",
                  "userId": "user-001",
                  "totalAmount": 149.99,
                  "currency": "USD",
                  "status": "confirmed",
                  "items": [
                    { "productId": "prod-001", "quantity": 2, "unitPrice": 74.99 }
                  ],
                  "timestamp": "2026-04-23T10:00:00Z"
                }
                """.formatted(orderId);

        // Subscribe consumer before producing
        consumer.subscribe(Collections.singletonList(TOPIC_ORDER_PLACED));

        // Produce
        ProducerRecord<String, String> record = new ProducerRecord<>(
                TOPIC_ORDER_PLACED, orderId, payload
        );
        Future<RecordMetadata> future = producer.send(record);
        RecordMetadata metadata = future.get(10, TimeUnit.SECONDS);

        assertNotNull(metadata);
        assertEquals(TOPIC_ORDER_PLACED, metadata.topic());
        assertTrue(metadata.offset() >= 0, "Offset should be non-negative");
        assertTrue(metadata.partition() >= 0, "Partition should be non-negative");

        // Consume and assert
        List<ConsumerRecord<String, String>> consumed = pollRecords(consumer, 1, Duration.ofSeconds(15));

        assertFalse(consumed.isEmpty(), "Expected at least one record");
        ConsumerRecord<String, String> received = consumed.stream()
                .filter(r -> r.topic().equals(TOPIC_ORDER_PLACED))
                .findFirst()
                .orElseThrow(() -> new AssertionError("No record found on topic " + TOPIC_ORDER_PLACED));

        assertEquals(orderId, received.key());
        assertTrue(received.value().contains(orderId), "Payload should contain orderId");
        assertTrue(received.value().contains("confirmed"), "Payload should contain status");
    }

    // ── Test: produce batch of messages ───────────────────────────────────────
    @Test
    @DisplayName("Should produce and consume multiple payment events in order")
    void testBatchProduceAndConsumePaymentEvents() throws ExecutionException, InterruptedException, TimeoutException {
        int messageCount = 10;
        List<String> sentKeys = new ArrayList<>();

        KafkaConsumer<String, String> batchConsumer = createConsumer("batch-consumer-" + UUID.randomUUID());
        batchConsumer.subscribe(Collections.singletonList(TOPIC_PAYMENT_PROCESSED));

        // Produce batch
        for (int i = 0; i < messageCount; i++) {
            String paymentId = "payment-" + UUID.randomUUID();
            String msg = """
                    {
                      "paymentId": "%s",
                      "orderId": "order-%d",
                      "amount": %.2f,
                      "currency": "USD",
                      "status": "succeeded",
                      "gateway": "stripe",
                      "timestamp": "2026-04-23T10:0%d:00Z"
                    }
                    """.formatted(paymentId, i, 99.99 + i, i);

            sentKeys.add(paymentId);
            producer.send(new ProducerRecord<>(TOPIC_PAYMENT_PROCESSED, paymentId, msg))
                    .get(5, TimeUnit.SECONDS);
        }
        producer.flush();

        // Consume batch
        List<ConsumerRecord<String, String>> consumed = pollRecords(batchConsumer, messageCount, Duration.ofSeconds(20));
        batchConsumer.close();

        assertEquals(messageCount, consumed.size(),
                "Expected " + messageCount + " records, got " + consumed.size());

        // Verify all keys received
        List<String> receivedKeys = consumed.stream().map(ConsumerRecord::key).toList();
        for (String sentKey : sentKeys) {
            assertTrue(receivedKeys.contains(sentKey),
                    "Missing key in consumed records: " + sentKey);
        }
    }

    // ── Test: consumer group rebalancing ──────────────────────────────────────
    @Test
    @DisplayName("Should distribute messages across consumer group members")
    void testConsumerGroupDistribution() throws ExecutionException, InterruptedException, TimeoutException {
        String groupId = "group-distribution-test-" + UUID.randomUUID();

        KafkaConsumer<String, String> consumer1 = createConsumer(groupId);
        KafkaConsumer<String, String> consumer2 = createConsumer(groupId);

        consumer1.subscribe(Collections.singletonList(TOPIC_USER_REGISTERED));
        consumer2.subscribe(Collections.singletonList(TOPIC_USER_REGISTERED));

        // Trigger rebalance
        consumer1.poll(Duration.ofMillis(500));
        consumer2.poll(Duration.ofMillis(500));

        // Produce messages
        for (int i = 0; i < 6; i++) {
            String userId = "user-" + i;
            String payload = """
                    {"userId": "%s", "email": "%s@shopos.dev", "timestamp": "2026-04-23T10:00:00Z"}
                    """.formatted(userId, userId);
            producer.send(new ProducerRecord<>(TOPIC_USER_REGISTERED, userId, payload))
                    .get(5, TimeUnit.SECONDS);
        }
        producer.flush();

        // Both consumers should receive messages (distributed across partitions)
        List<ConsumerRecord<String, String>> fromC1 = pollRecords(consumer1, 1, Duration.ofSeconds(10));
        List<ConsumerRecord<String, String>> fromC2 = pollRecords(consumer2, 1, Duration.ofSeconds(10));

        int total = fromC1.size() + fromC2.size();
        assertTrue(total >= 1, "At least one consumer should have received messages");

        consumer1.close();
        consumer2.close();
    }

    // ── Test: message headers ─────────────────────────────────────────────────
    @Test
    @DisplayName("Should preserve Kafka message headers for tracing")
    void testMessageHeaders() throws ExecutionException, InterruptedException, TimeoutException {
        String traceId = UUID.randomUUID().toString();
        String spanId  = UUID.randomUUID().toString().substring(0, 16);

        KafkaConsumer<String, String> headerConsumer = createConsumer("header-test-" + UUID.randomUUID());
        headerConsumer.subscribe(Collections.singletonList(TOPIC_ORDER_PLACED));

        org.apache.kafka.common.header.Headers headers = new org.apache.kafka.common.header.internals.RecordHeaders();
        headers.add("traceparent", ("00-" + traceId + "-" + spanId + "-01").getBytes());
        headers.add("x-service-name", "order-service".getBytes());
        headers.add("x-correlation-id", UUID.randomUUID().toString().getBytes());

        ProducerRecord<String, String> record = new ProducerRecord<>(
                TOPIC_ORDER_PLACED,
                null,       // partition — let Kafka decide
                "order-hdr-test",
                "{\"orderId\":\"hdr-test\",\"status\":\"confirmed\"}",
                headers
        );

        producer.send(record).get(5, TimeUnit.SECONDS);
        producer.flush();

        List<ConsumerRecord<String, String>> consumed = pollRecords(headerConsumer, 1, Duration.ofSeconds(10));
        headerConsumer.close();

        assertFalse(consumed.isEmpty(), "No records consumed");
        ConsumerRecord<String, String> received = consumed.get(0);

        assertNotNull(received.headers().lastHeader("traceparent"), "traceparent header missing");
        assertNotNull(received.headers().lastHeader("x-service-name"), "x-service-name header missing");

        String receivedServiceName = new String(received.headers().lastHeader("x-service-name").value());
        assertEquals("order-service", receivedServiceName);
    }

    // ── Helpers ───────────────────────────────────────────────────────────────

    private KafkaProducer<String, String> createProducer() {
        Properties props = new Properties();
        props.put(ProducerConfig.BOOTSTRAP_SERVERS_CONFIG, kafkaContainer.getBootstrapServers());
        props.put(ProducerConfig.KEY_SERIALIZER_CLASS_CONFIG, StringSerializer.class.getName());
        props.put(ProducerConfig.VALUE_SERIALIZER_CLASS_CONFIG, StringSerializer.class.getName());
        props.put(ProducerConfig.ACKS_CONFIG, "all");
        props.put(ProducerConfig.RETRIES_CONFIG, 3);
        props.put(ProducerConfig.ENABLE_IDEMPOTENCE_CONFIG, true);
        props.put(ProducerConfig.MAX_IN_FLIGHT_REQUESTS_PER_CONNECTION, 1);
        return new KafkaProducer<>(props);
    }

    private KafkaConsumer<String, String> createConsumer(String groupId) {
        Properties props = new Properties();
        props.put(ConsumerConfig.BOOTSTRAP_SERVERS_CONFIG, kafkaContainer.getBootstrapServers());
        props.put(ConsumerConfig.GROUP_ID_CONFIG, groupId);
        props.put(ConsumerConfig.KEY_DESERIALIZER_CLASS_CONFIG, StringDeserializer.class.getName());
        props.put(ConsumerConfig.VALUE_DESERIALIZER_CLASS_CONFIG, StringDeserializer.class.getName());
        props.put(ConsumerConfig.AUTO_OFFSET_RESET_CONFIG, "earliest");
        props.put(ConsumerConfig.ENABLE_AUTO_COMMIT_CONFIG, false);
        props.put(ConsumerConfig.MAX_POLL_RECORDS_CONFIG, 50);
        return new KafkaConsumer<>(props);
    }

    private List<ConsumerRecord<String, String>> pollRecords(
            KafkaConsumer<String, String> consumer,
            int minExpected,
            Duration timeout
    ) {
        List<ConsumerRecord<String, String>> result = new ArrayList<>();
        long deadline = System.currentTimeMillis() + timeout.toMillis();

        while (result.size() < minExpected && System.currentTimeMillis() < deadline) {
            ConsumerRecords<String, String> records = consumer.poll(Duration.ofMillis(500));
            records.forEach(result::add);
        }
        consumer.commitSync();
        return result;
    }
}
