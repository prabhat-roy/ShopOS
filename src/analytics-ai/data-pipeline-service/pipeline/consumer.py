import asyncio
import json
import logging
from datetime import datetime, timezone
from typing import Optional

from aiokafka import AIOKafkaConsumer, AIOKafkaProducer

from pipeline.config import settings
from pipeline.models import RawEvent
from pipeline.store import PipelineStore
from pipeline.transformer import EventTransformer

logger = logging.getLogger(__name__)


class KafkaPipelineConsumer:
    def __init__(self, store: PipelineStore) -> None:
        self._store = store
        self._transformer = EventTransformer()
        self._consumer: Optional[AIOKafkaConsumer] = None
        self._producer: Optional[AIOKafkaProducer] = None
        self._running = False

    async def start(self) -> None:
        self._consumer = AIOKafkaConsumer(
            *settings.input_topics_list,
            bootstrap_servers=settings.KAFKA_BROKERS,
            group_id=settings.KAFKA_GROUP_ID,
            auto_offset_reset="earliest",
            enable_auto_commit=True,
            value_deserializer=lambda m: m.decode("utf-8", errors="replace"),
            key_deserializer=lambda k: k.decode("utf-8", errors="replace") if k else None,
        )
        self._producer = AIOKafkaProducer(
            bootstrap_servers=settings.KAFKA_BROKERS,
            value_serializer=lambda v: json.dumps(v).encode("utf-8"),
            key_serializer=lambda k: k.encode("utf-8") if k else None,
        )
        await self._consumer.start()
        await self._producer.start()
        self._running = True
        logger.info(
            "Kafka consumer started. Listening on topics: %s",
            settings.INPUT_TOPICS,
        )

    async def stop(self) -> None:
        self._running = False
        if self._consumer:
            await self._consumer.stop()
        if self._producer:
            await self._producer.stop()
        logger.info("Kafka consumer stopped.")

    async def consume(self) -> None:
        if not self._consumer or not self._producer:
            raise RuntimeError("Consumer not started. Call start() first.")

        async for msg in self._consumer:
            if not self._running:
                break
            await self._process_message(msg.topic, msg.value)

    async def _process_message(self, topic: str, raw_value: str) -> None:
        await self._store.increment_processed()
        try:
            payload = json.loads(raw_value)
            raw_event = RawEvent(
                topic=topic,
                eventId=payload.get("eventId"),
                data=payload,
                receivedAt=datetime.now(timezone.utc),
            )
            enriched = self._transformer.transform(raw_event)
            await self._store.save_event(enriched)

            if self._producer:
                enriched_payload = {
                    "eventId": enriched.eventId,
                    "topic": enriched.topic,
                    "originalData": enriched.originalData,
                    "enrichedData": enriched.enrichedData,
                    "transformedAt": enriched.transformedAt.isoformat(),
                    "processingTimeMs": enriched.processingTimeMs,
                }
                await self._producer.send_and_wait(
                    settings.OUTPUT_TOPIC,
                    value=enriched_payload,
                    key=enriched.eventId,
                )

        except json.JSONDecodeError as exc:
            logger.warning("Failed to decode JSON from topic %s: %s", topic, exc)
            await self._store.increment_failed()
        except Exception as exc:
            logger.error("Unexpected error processing message from topic %s: %s", topic, exc, exc_info=True)
            await self._store.increment_failed()
