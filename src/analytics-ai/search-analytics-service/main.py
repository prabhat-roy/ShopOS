"""
search-analytics-service — entry point.

HTTP  : port 8196 (FastAPI / uvicorn)
Kafka : consumer group 'search-analytics-service'
DB    : Cassandra (search event storage)
"""

from __future__ import annotations

import asyncio
import json
import logging
import sys
from typing import Optional

import uvicorn
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from searchanalytics.api import router
from searchanalytics.config import settings

# ---------------------------------------------------------------------------
# Logging
# ---------------------------------------------------------------------------

logging.basicConfig(
    level=getattr(logging, settings.LOG_LEVEL.upper(), logging.INFO),
    format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
    stream=sys.stdout,
)
logger = logging.getLogger(__name__)

# ---------------------------------------------------------------------------
# Application
# ---------------------------------------------------------------------------

app = FastAPI(
    title="Search Analytics Service",
    description=(
        "Tracks search terms, zero-result searches, click-through rates, and "
        "search-to-purchase conversion. Powers search quality dashboards and "
        "auto-complete tuning."
    ),
    version="1.0.0",
    docs_url="/docs",
    redoc_url="/redoc",
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(router)

# ---------------------------------------------------------------------------
# Kafka consumer (background task)
# ---------------------------------------------------------------------------

_kafka_task: Optional[asyncio.Task] = None


async def _kafka_consumer_loop() -> None:
    """Consume search analytics events from Kafka."""
    try:
        from aiokafka import AIOKafkaConsumer  # type: ignore[import]
    except ImportError:
        logger.warning("aiokafka not installed — Kafka consumer disabled")
        return

    topics = [t.strip() for t in settings.KAFKA_TOPICS.split(",") if t.strip()]

    consumer = AIOKafkaConsumer(
        *topics,
        bootstrap_servers=settings.KAFKA_BROKERS,
        group_id=settings.KAFKA_GROUP_ID,
        auto_offset_reset="latest",
        enable_auto_commit=True,
        value_deserializer=lambda v: json.loads(v.decode("utf-8", errors="replace")),
    )

    try:
        await consumer.start()
        logger.info("Kafka consumer started — topics: %s", topics)
        async for msg in consumer:
            try:
                payload = msg.value
                logger.debug(
                    "Received search event topic=%s query=%s",
                    msg.topic,
                    payload.get("query", ""),
                )
                # Real implementation: store in Cassandra search_events table
            except Exception as exc:
                logger.warning("Error processing Kafka message: %s", exc)
    except Exception as exc:
        logger.error("Kafka consumer error: %s", exc)
    finally:
        try:
            await consumer.stop()
        except Exception:
            pass
        logger.info("Kafka consumer stopped")


# ---------------------------------------------------------------------------
# Startup / shutdown
# ---------------------------------------------------------------------------


@app.on_event("startup")
async def startup() -> None:
    global _kafka_task
    _kafka_task = asyncio.create_task(_kafka_consumer_loop())
    logger.info("search-analytics-service starting on HTTP :%d", settings.HTTP_PORT)


@app.on_event("shutdown")
async def shutdown() -> None:
    global _kafka_task
    if _kafka_task and not _kafka_task.done():
        _kafka_task.cancel()
        try:
            await _kafka_task
        except asyncio.CancelledError:
            pass
    logger.info("search-analytics-service shut down")


# ---------------------------------------------------------------------------
# Entry point
# ---------------------------------------------------------------------------

if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=settings.HTTP_PORT,
        log_level=settings.LOG_LEVEL.lower(),
        reload=False,
    )
