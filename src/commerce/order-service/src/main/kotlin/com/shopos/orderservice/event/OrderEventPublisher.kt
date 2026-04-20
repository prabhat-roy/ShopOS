package com.shopos.orderservice.event

import com.fasterxml.jackson.databind.ObjectMapper
import org.slf4j.LoggerFactory
import org.springframework.kafka.core.KafkaTemplate
import org.springframework.stereotype.Component
import java.util.UUID

@Component
class OrderEventPublisher(
    private val kafkaTemplate: KafkaTemplate<String, String>,
    private val objectMapper: ObjectMapper
) {

    private val log = LoggerFactory.getLogger(OrderEventPublisher::class.java)

    companion object {
        const val TOPIC_ORDER_PLACED    = "commerce.order.placed"
        const val TOPIC_ORDER_CANCELLED = "commerce.order.cancelled"
    }

    fun publishOrderPlaced(orderId: UUID, payload: Any) {
        publish(TOPIC_ORDER_PLACED, orderId.toString(), payload)
    }

    fun publishOrderCancelled(orderId: UUID, payload: Any) {
        publish(TOPIC_ORDER_CANCELLED, orderId.toString(), payload)
    }

    fun publish(topic: String, key: String, payload: Any) {
        val json = objectMapper.writeValueAsString(payload)
        kafkaTemplate.send(topic, key, json)
            .whenComplete { result, ex ->
                if (ex != null) {
                    log.error("Failed to publish event to topic={} key={}: {}", topic, key, ex.message)
                } else {
                    log.debug(
                        "Published event to topic={} key={} offset={}",
                        topic, key, result.recordMetadata.offset()
                    )
                }
            }
    }
}
