from __future__ import annotations

import json
import logging
from typing import List, Optional

from aiokafka import AIOKafkaProducer
from aiokafka.errors import KafkaError

from tracker.config import settings
from tracker.models import EventType, TrackingEvent

logger = logging.getLogger(__name__)

_TOPIC_MAP = {
    EventType.PAGE_VIEW: settings.KAFKA_TOPIC_ANALYTICS,
    EventType.CLICK: "analytics.product.clicked",
    EventType.IMPRESSION: "analytics.product.clicked",
    EventType.CONVERSION: "analytics.product.clicked",
    EventType.CUSTOM: settings.KAFKA_TOPIC_ANALYTICS,
}


class KafkaEventPublisher:
    """Publishes tracking events to appropriate Kafka topics."""

    def __init__(self) -> None:
        self._producer: Optional[AIOKafkaProducer] = None

    async def start(self) -> None:
        self._producer = AIOKafkaProducer(
            bootstrap_servers=settings.KAFKA_BROKERS,
            value_serializer=lambda v: json.dumps(v).encode("utf-8"),
            acks="all",
            enable_idempotence=True,
        )
        await self._producer.start()
        logger.info("Kafka producer started. Brokers: %s", settings.KAFKA_BROKERS)

    async def stop(self) -> None:
        if self._producer:
            await self._producer.stop()
            logger.info("Kafka producer stopped.")

    async def publish(self, event_id: str, event: TrackingEvent) -> bool:
        if self._producer is None:
            logger.warning("Producer not running — skipping publish for event %s", event_id)
            return False

        topic = _TOPIC_MAP.get(event.eventType, settings.KAFKA_TOPIC_ANALYTICS)
        payload = {
            "eventId": event_id,
            "eventType": event.eventType.value,
            "sessionId": event.sessionId,
            "userId": event.userId,
            "data": event.data,
            "clientTimestamp": event.clientTimestamp.isoformat() if event.clientTimestamp else None,
            "receivedAt": event.receivedAt.isoformat(),
            "ip": event.ip,
        }

        try:
            await self._producer.send_and_wait(topic, payload)
            logger.debug("Published event %s to topic %s", event_id, topic)
            return True
        except KafkaError as exc:
            logger.error("Failed to publish event %s: %s", event_id, exc)
            return False

    async def batch_publish(self, events: List[tuple[str, TrackingEvent]]) -> List[bool]:
        results = []
        for event_id, event in events:
            result = await self.publish(event_id, event)
            results.append(result)
        return results
