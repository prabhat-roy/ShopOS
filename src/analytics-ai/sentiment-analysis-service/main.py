"""
sentiment-analysis-service — entry point.

HTTP  : port 8703 (FastAPI / uvicorn)
Kafka : consumer group 'sentiment-analysis'
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

from sentiment.api import router
from sentiment.config import settings

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
    title="Sentiment Analysis Service",
    description=(
        "Rule-based NLP sentiment analysis for product reviews and customer feedback. "
        "Persists results to PostgreSQL; optionally consumes Kafka events."
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
    """
    Consume messages from Kafka topics and analyse any text payload found.

    The consumer is optional: if Kafka is unavailable the service continues
    to function via the HTTP API.
    """
    try:
        from aiokafka import AIOKafkaConsumer  # type: ignore[import]
    except ImportError:
        logger.warning("aiokafka not installed — Kafka consumer disabled")
        return

    from sentiment.analyzer import analyzer

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
                # Extract text from known event shapes
                text: Optional[str] = (
                    payload.get("review")
                    or payload.get("comment")
                    or payload.get("feedback")
                    or payload.get("text")
                    or payload.get("body")
                )
                if text:
                    entity_id = payload.get("reviewId") or payload.get("id")
                    entity_type = payload.get("entityType", "review")
                    result = analyzer.analyze(
                        text=text,
                        entity_id=entity_id,
                        entity_type=entity_type,
                    )
                    store = app.state.store
                    if store is not None:
                        await store.save_result(result)
                    logger.debug(
                        "Kafka msg processed: entity_id=%s label=%s",
                        entity_id,
                        result.label,
                    )
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

    # Connect to PostgreSQL
    try:
        from sentiment.store import AsyncPgStore

        store = await AsyncPgStore.create(settings.DATABASE_URL)
        app.state.store = store
        logger.info("PostgreSQL connection pool ready")
    except Exception as exc:
        logger.warning("PostgreSQL unavailable — running without persistence: %s", exc)
        app.state.store = None

    # Start Kafka consumer in background
    _kafka_task = asyncio.create_task(_kafka_consumer_loop())

    logger.info(
        "sentiment-analysis-service starting on HTTP :%d",
        settings.HTTP_PORT,
    )


@app.on_event("shutdown")
async def shutdown() -> None:
    global _kafka_task

    if _kafka_task and not _kafka_task.done():
        _kafka_task.cancel()
        try:
            await _kafka_task
        except asyncio.CancelledError:
            pass

    store = getattr(app.state, "store", None)
    if store is not None:
        await store.close()
        logger.info("PostgreSQL connection pool closed")

    logger.info("sentiment-analysis-service shut down")


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
