package com.shopos.paymentservice.event;

import com.shopos.paymentservice.domain.Payment;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Component;

import java.util.Map;

@Slf4j
@Component
@RequiredArgsConstructor
public class PaymentEventPublisher {

    static final String TOPIC_PROCESSED = "commerce.payment.processed";
    static final String TOPIC_FAILED    = "commerce.payment.failed";

    private final KafkaTemplate<String, Object> kafkaTemplate;

    public void publishProcessed(Payment payment) {
        Map<String, Object> event = buildPayload(payment, "PROCESSED");
        log.info("Publishing {} for paymentId={}", TOPIC_PROCESSED, payment.getId());
        kafkaTemplate.send(TOPIC_PROCESSED, payment.getId().toString(), event);
    }

    public void publishFailed(Payment payment, String reason) {
        Map<String, Object> event = buildPayload(payment, "FAILED");
        event.put("reason", reason);
        log.warn("Publishing {} for paymentId={} reason={}", TOPIC_FAILED, payment.getId(), reason);
        kafkaTemplate.send(TOPIC_FAILED, payment.getId().toString(), event);
    }

    private Map<String, Object> buildPayload(Payment payment, String eventType) {
        // Using a mutable map so callers can add extra fields (e.g. reason)
        java.util.Map<String, Object> payload = new java.util.HashMap<>();
        payload.put("eventType",   eventType);
        payload.put("paymentId",   payment.getId().toString());
        payload.put("orderId",     payment.getOrderId());
        payload.put("customerId",  payment.getCustomerId());
        payload.put("amount",      payment.getAmount());
        payload.put("currency",    payment.getCurrency());
        payload.put("status",      payment.getStatus().name());
        payload.put("provider",    payment.getProvider());
        payload.put("occurredAt",  java.time.Instant.now().toString());
        return payload;
    }
}
