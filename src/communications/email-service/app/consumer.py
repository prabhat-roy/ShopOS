from __future__ import annotations

import asyncio
import json
import logging
from typing import Optional

from aiokafka import AIOKafkaConsumer
from aiokafka.errors import KafkaError

from .config import settings
from .models import EmailMessage
from .sender import email_sender
from .store import email_store

logger = logging.getLogger(__name__)


class KafkaEmailConsumer:
    """Async Kafka consumer that processes ``email.send`` topic messages.

    Design decisions:
    - Manual offset commit: only commit after a record is successfully
      persisted so that transient failures trigger reprocessing.
    - Error isolation: JSON decode errors and validation errors are logged
      and skipped; the consumer keeps running.
    - The running state is exposed so that the health endpoint can surface it.
    """

    def __init__(self) -> None:
        self._consumer: Optional[AIOKafkaConsumer] = None
        self._running: bool = False
        self._task: Optional[asyncio.Task] = None

    @property
    def is_running(self) -> bool:
        return self._running and self._task is not None and not self._task.done()

    async def start(self) -> None:
        """Create and start the Kafka consumer, then launch the poll loop."""
        self._consumer = AIOKafkaConsumer(
            settings.KAFKA_TOPIC,
            bootstrap_servers=settings.KAFKA_BROKERS,
            group_id=settings.KAFKA_GROUP_ID,
            auto_offset_reset=settings.KAFKA_AUTO_OFFSET_RESET,
            enable_auto_commit=False,
            session_timeout_ms=settings.KAFKA_SESSION_TIMEOUT_MS,
            heartbeat_interval_ms=settings.KAFKA_HEARTBEAT_INTERVAL_MS,
            value_deserializer=lambda v: v,  # raw bytes; we decode manually
        )
        await self._consumer.start()
        self._running = True
        self._task = asyncio.create_task(self._poll_loop(), name="kafka-email-consumer")
        logger.info(
            "KafkaEmailConsumer started — brokers=%s topic=%s group=%s",
            settings.KAFKA_BROKERS,
            settings.KAFKA_TOPIC,
            settings.KAFKA_GROUP_ID,
        )

    async def stop(self) -> None:
        """Gracefully stop the poll loop and close the consumer."""
        self._running = False
        if self._task and not self._task.done():
            self._task.cancel()
            try:
                await self._task
            except asyncio.CancelledError:
                pass
        if self._consumer:
            await self._consumer.stop()
            logger.info("KafkaEmailConsumer stopped")

    # ------------------------------------------------------------------
    # Internal
    # ------------------------------------------------------------------

    async def _poll_loop(self) -> None:
        """Continuously fetch messages from Kafka and process them."""
        assert self._consumer is not None

        while self._running:
            try:
                async for message in self._consumer:
                    if not self._running:
                        break
                    await self._handle_message(message)
            except asyncio.CancelledError:
                logger.info("KafkaEmailConsumer poll loop cancelled")
                break
            except KafkaError as exc:
                logger.error("Kafka error in poll loop: %s — reconnecting in 5 s", exc)
                await asyncio.sleep(5)
            except Exception as exc:  # pragma: no cover
                logger.exception("Unexpected error in poll loop: %s", exc)
                await asyncio.sleep(2)

    async def _handle_message(self, message) -> None:
        """Parse, process, and commit a single Kafka message."""
        raw_value: bytes = message.value
        topic = message.topic
        partition = message.partition
        offset = message.offset

        log_ctx = f"topic={topic} partition={partition} offset={offset}"

        # --- Step 1: JSON decode ----------------------------------------
        try:
            payload = json.loads(raw_value.decode("utf-8"))
        except (json.JSONDecodeError, UnicodeDecodeError) as exc:
            logger.warning("Skipping undecodable message (%s): %s", log_ctx, exc)
            # Commit anyway so we don't block the consumer indefinitely
            await self._consumer.commit()
            return

        # --- Step 2: Pydantic validation ---------------------------------
        try:
            email_msg = EmailMessage.model_validate(payload)
        except Exception as exc:
            logger.warning("Skipping invalid EmailMessage (%s): %s", log_ctx, exc)
            await self._consumer.commit()
            return

        logger.debug(
            "Processing email messageId=%s to=%s (%s)",
            email_msg.messageId,
            email_msg.to,
            log_ctx,
        )

        # --- Step 3: Send ------------------------------------------------
        try:
            record = email_sender.send(email_msg)
        except Exception as exc:
            logger.error("Sender raised unexpected error for %s: %s", email_msg.messageId, exc)
            # Do NOT commit; allow retry
            return

        # --- Step 4: Persist ---------------------------------------------
        try:
            await email_store.save(record)
        except Exception as exc:
            logger.error("Store.save failed for %s: %s", email_msg.messageId, exc)
            return

        # --- Step 5: Commit offset ---------------------------------------
        await self._consumer.commit()
        logger.info(
            "Committed offset %s — messageId=%s status=%s",
            offset,
            record.messageId,
            record.status,
        )


# Module-level singleton
kafka_consumer = KafkaEmailConsumer()
