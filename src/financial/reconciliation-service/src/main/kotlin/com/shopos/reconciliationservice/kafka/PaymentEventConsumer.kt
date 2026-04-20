package com.shopos.reconciliationservice.kafka

import com.fasterxml.jackson.annotation.JsonIgnoreProperties
import com.fasterxml.jackson.databind.ObjectMapper
import com.shopos.reconciliationservice.dto.ReconcileRequest
import com.shopos.reconciliationservice.service.ReconciliationService
import org.apache.kafka.clients.consumer.ConsumerRecord
import org.slf4j.LoggerFactory
import org.springframework.kafka.annotation.KafkaListener
import org.springframework.kafka.support.Acknowledgment
import org.springframework.stereotype.Component
import java.math.BigDecimal
import java.util.UUID

/**
 * Payload shape expected on the `commerce.payment.processed` topic.
 * Fields marked as optional/nullable are treated as not-yet-reconciled if absent.
 */
@JsonIgnoreProperties(ignoreUnknown = true)
data class PaymentProcessedEvent(
    val paymentId: String,
    val amount: BigDecimal,
    val currency: String = "USD",
    val processor: String,
    val externalTransactionId: String? = null,
    val status: String = "PROCESSED"
)

@Component
class PaymentEventConsumer(
    private val reconciliationService: ReconciliationService,
    private val objectMapper: ObjectMapper
) {

    private val log = LoggerFactory.getLogger(PaymentEventConsumer::class.java)

    @KafkaListener(
        topics = ["commerce.payment.processed"],
        groupId = "\${spring.kafka.consumer.group-id:reconciliation-service-group}",
        containerFactory = "kafkaListenerContainerFactory"
    )
    fun onPaymentProcessed(record: ConsumerRecord<String, String>, ack: Acknowledgment) {
        log.debug(
            "Received payment event: topic={} partition={} offset={}",
            record.topic(), record.partition(), record.offset()
        )

        try {
            val event = objectMapper.readValue(record.value(), PaymentProcessedEvent::class.java)

            // Only auto-reconcile when the processor has supplied an external transaction ID
            if (event.externalTransactionId.isNullOrBlank()) {
                log.info(
                    "Skipping auto-reconcile for paymentId={}: no externalTransactionId present",
                    event.paymentId
                )
                ack.acknowledge()
                return
            }

            val request = ReconcileRequest(
                internalPaymentId     = UUID.fromString(event.paymentId),
                externalTransactionId = event.externalTransactionId,
                internalAmount        = event.amount,
                externalAmount        = event.amount,   // treat as matched when only one amount is reported
                currency              = event.currency,
                processor             = event.processor
            )

            val response = reconciliationService.reconcile(request)
            log.info(
                "Auto-reconciled paymentId={} → recordId={} status={}",
                event.paymentId, response.id, response.status
            )

        } catch (ex: IllegalStateException) {
            // Duplicate — already reconciled; acknowledge to avoid reprocessing
            log.warn("Duplicate reconciliation skipped: {}", ex.message)
        } catch (ex: Exception) {
            log.error(
                "Failed to process payment event at offset={}: {}",
                record.offset(), ex.message, ex
            )
            // Re-throw to let the error handler / DLQ deal with it
            throw ex
        }

        ack.acknowledge()
    }
}
