from __future__ import annotations

import json
import logging
from datetime import date, datetime

from aiokafka import AIOKafkaConsumer

from forecast.config import settings
from forecast.models import SalesRecord
from forecast.store import AsyncPgStore

logger = logging.getLogger(__name__)


class KafkaOrderConsumer:
    """
    Consumes commerce.order.placed events from Kafka and persists each
    line-item quantity as a SalesRecord for downstream forecasting.

    Expected event shape (JSON):
    {
        "orderId": "ord-123",
        "placedAt": "2024-03-15T12:00:00Z",  // ISO-8601 or YYYY-MM-DD
        "lineItems": [
            {
                "productId": "prod-456",
                "sku": "SKU-001",
                "quantity": 3
            },
            ...
        ]
    }
    """

    def __init__(self, store: AsyncPgStore) -> None:
        self._store = store
        self._consumer: AIOKafkaConsumer | None = None

    async def start(self) -> None:
        self._consumer = AIOKafkaConsumer(
            settings.KAFKA_TOPIC,
            bootstrap_servers=settings.KAFKA_BROKERS,
            group_id=settings.KAFKA_GROUP_ID,
            auto_offset_reset="earliest",
            enable_auto_commit=False,
            value_deserializer=lambda raw: json.loads(raw.decode("utf-8")),
        )
        await self._consumer.start()
        logger.info(
            "Kafka consumer started. topic=%s group=%s brokers=%s",
            settings.KAFKA_TOPIC,
            settings.KAFKA_GROUP_ID,
            settings.KAFKA_BROKERS,
        )

    async def stop(self) -> None:
        if self._consumer:
            await self._consumer.stop()
            logger.info("Kafka consumer stopped.")

    async def consume(self) -> None:
        """Main consume loop — runs until cancelled."""
        if self._consumer is None:
            raise RuntimeError("Consumer not started. Call start() first.")

        async for msg in self._consumer:
            try:
                await self._process_message(msg.value)
                await self._consumer.commit()
            except Exception:
                logger.exception(
                    "Failed to process Kafka message. offset=%s partition=%s",
                    msg.offset,
                    msg.partition,
                )
                # Skip the message after logging — do NOT commit so that
                # on a restart the broker re-delivers from last committed offset.
                # For production, a dead-letter topic should be used instead.

    async def _process_message(self, payload: dict) -> None:
        order_id: str = payload.get("orderId", "")
        if not order_id:
            logger.warning("Received order event without orderId: %s", payload)
            return

        raw_date = payload.get("placedAt") or payload.get("orderDate") or ""
        sale_date = _parse_date(raw_date)

        line_items: list[dict] = payload.get("lineItems", [])
        if not line_items:
            logger.debug("Order %s has no line items; skipping.", order_id)
            return

        for item in line_items:
            product_id: str = item.get("productId", "")
            sku: str = item.get("sku", "")
            quantity: int = int(item.get("quantity", 0))

            if not product_id or not sku or quantity <= 0:
                logger.warning(
                    "Skipping invalid line item in order %s: %s", order_id, item
                )
                continue

            record = SalesRecord(
                productId=product_id,
                sku=sku,
                quantity=quantity,
                saleDate=sale_date,
                orderId=order_id,
            )
            await self._store.save_sale(record)
            logger.debug(
                "Saved sale record. orderId=%s productId=%s sku=%s qty=%d date=%s",
                order_id,
                product_id,
                sku,
                quantity,
                sale_date,
            )


def _parse_date(raw: str) -> date:
    """Parse ISO-8601 datetime or date string; fall back to today on failure."""
    if not raw:
        return date.today()
    try:
        if "T" in raw or " " in raw:
            return datetime.fromisoformat(raw.replace("Z", "+00:00")).date()
        return date.fromisoformat(raw)
    except (ValueError, AttributeError):
        logger.warning("Could not parse date '%s'; defaulting to today.", raw)
        return date.today()
