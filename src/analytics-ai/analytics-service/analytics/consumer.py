from __future__ import annotations

import json
import logging
from typing import Optional

from aiokafka import AIOKafkaConsumer
from aiokafka.errors import KafkaError

from analytics.config import settings
from analytics.models import PageViewEvent, ProductClickEvent, SearchEvent
from analytics.store import AnalyticsStore

logger = logging.getLogger(__name__)

TOPIC_PAGE_VIEWED = "analytics.page.viewed"
TOPIC_PRODUCT_CLICKED = "analytics.product.clicked"
TOPIC_SEARCH_PERFORMED = "analytics.search.performed"


class KafkaAnalyticsConsumer:
    """Subscribes to analytics Kafka topics and persists events to the store."""

    def __init__(self, store: AnalyticsStore) -> None:
        self._store = store
        self._consumer: Optional[AIOKafkaConsumer] = None
        self._running = False

    async def start(self) -> None:
        self._consumer = AIOKafkaConsumer(
            *settings.kafka_topic_list,
            bootstrap_servers=settings.KAFKA_BROKERS,
            group_id=settings.KAFKA_GROUP_ID,
            enable_auto_commit=False,
            auto_offset_reset="earliest",
            value_deserializer=lambda v: v.decode("utf-8"),
        )
        await self._consumer.start()
        self._running = True
        logger.info(
            "Kafka consumer started. Topics: %s, Brokers: %s",
            settings.kafka_topic_list,
            settings.KAFKA_BROKERS,
        )

    async def stop(self) -> None:
        self._running = False
        if self._consumer:
            await self._consumer.stop()
            logger.info("Kafka consumer stopped.")

    async def consume(self) -> None:
        if self._consumer is None:
            raise RuntimeError("Consumer not started — call start() first.")

        async for msg in self._consumer:
            if not self._running:
                break
            try:
                await self._handle_message(msg.topic, msg.value)
                await self._consumer.commit()
            except KafkaError as exc:
                logger.error("Kafka error on topic %s: %s", msg.topic, exc)
            except Exception as exc:  # noqa: BLE001
                logger.error(
                    "Unhandled error processing message from topic %s: %s",
                    msg.topic,
                    exc,
                    exc_info=True,
                )
                # Commit so we don't replay bad messages indefinitely.
                try:
                    await self._consumer.commit()
                except Exception:  # noqa: BLE001
                    pass

    async def _handle_message(self, topic: str, raw: str) -> None:
        payload = json.loads(raw)

        if topic == TOPIC_PAGE_VIEWED:
            event = PageViewEvent.model_validate(payload)
            await self._store.save_page_view(event)
            logger.debug("Saved page_view for session=%s url=%s", event.sessionId, event.pageUrl)

        elif topic == TOPIC_PRODUCT_CLICKED:
            event = ProductClickEvent.model_validate(payload)
            await self._store.save_product_click(event)
            logger.debug(
                "Saved product_click for session=%s product=%s",
                event.sessionId,
                event.productId,
            )

        elif topic == TOPIC_SEARCH_PERFORMED:
            event = SearchEvent.model_validate(payload)
            await self._store.save_search(event)
            logger.debug(
                "Saved search for session=%s query=%s",
                event.sessionId,
                event.query,
            )

        else:
            logger.warning("Received message on unhandled topic: %s", topic)
