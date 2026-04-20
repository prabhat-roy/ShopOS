"""
demand-forecast-service
=======================
Starts the Kafka consumer as a FastAPI background task and exposes the
forecast REST API on HTTP_PORT (default 8209).
"""

from __future__ import annotations

import asyncio
import logging
import sys

import uvicorn
from fastapi import FastAPI

from forecast.api import router
from forecast.config import settings
from forecast.consumer import KafkaOrderConsumer
from forecast.engine import ForecastEngine
from forecast.store import AsyncPgStore

logging.basicConfig(
    stream=sys.stdout,
    level=logging.INFO,
    format="%(asctime)s %(levelname)-8s %(name)s %(message)s",
)
logger = logging.getLogger(__name__)

app = FastAPI(
    title="Demand Forecast Service",
    description="Consumes order events and produces demand forecasts via moving average.",
    version="1.0.0",
)

app.include_router(router)


@app.on_event("startup")
async def startup() -> None:
    # Initialise store
    store = AsyncPgStore(settings.DATABASE_URL)
    try:
        await store.init()
        app.state.store = store
    except Exception as exc:
        logger.warning("PostgreSQL unavailable — running without persistence: %s", exc)
        app.state.store = None
    app.state.engine = ForecastEngine()

    # Start Kafka consumer in background
    try:
        consumer = KafkaOrderConsumer(store)
        await consumer.start()
        app.state.consumer = consumer
        loop = asyncio.get_event_loop()
        app.state.consumer_task = loop.create_task(consumer.consume())
    except Exception as exc:
        logger.warning("Kafka unavailable — consumer not started: %s", exc)
        app.state.consumer = None
        app.state.consumer_task = None
    logger.info("Demand Forecast Service started. port=%d", settings.HTTP_PORT)


@app.on_event("shutdown")
async def shutdown() -> None:
    task: asyncio.Task | None = getattr(app.state, "consumer_task", None)
    if task is not None:
        task.cancel()
        try:
            await task
        except asyncio.CancelledError:
            pass

    consumer = getattr(app.state, "consumer", None)
    if consumer is not None:
        await consumer.stop()

    store = getattr(app.state, "store", None)
    if store is not None:
        await store.close()
    logger.info("Demand Forecast Service shut down cleanly.")


if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=settings.HTTP_PORT,
        log_config=None,
    )
