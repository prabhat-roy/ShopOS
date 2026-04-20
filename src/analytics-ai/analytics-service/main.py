from __future__ import annotations

import asyncio
import logging
import signal
from contextlib import asynccontextmanager
from typing import AsyncIterator

import uvicorn
from fastapi import FastAPI

from analytics.api import router
from analytics.config import settings
from analytics.consumer import KafkaAnalyticsConsumer
from analytics.store import AnalyticsStore, CassandraAnalyticsStore, InMemoryAnalyticsStore

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(levelname)s %(name)s — %(message)s",
)
logger = logging.getLogger(__name__)


def _build_store() -> AnalyticsStore:
    if settings.STORE_BACKEND == "memory":
        logger.info("Using InMemoryAnalyticsStore (testing mode)")
        return InMemoryAnalyticsStore()

    try:
        from cassandra.cluster import Cluster  # type: ignore[import-untyped]

        cluster = Cluster(settings.cassandra_host_list, connect_timeout=2)
        session = cluster.connect()
        logger.info("Connected to Cassandra at %s", settings.cassandra_host_list)
        return CassandraAnalyticsStore(session, settings.CASSANDRA_KEYSPACE)
    except Exception as exc:  # noqa: BLE001
        logger.warning(
            "Cassandra unavailable (%s) — falling back to InMemoryAnalyticsStore", exc
        )
        return InMemoryAnalyticsStore()


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncIterator[None]:
    store = _build_store()
    app.state.store = store

    consumer = KafkaAnalyticsConsumer(store)
    consume_task: asyncio.Task | None = None

    try:
        await consumer.start()
        consume_task = asyncio.create_task(consumer.consume(), name="kafka-consumer")
        logger.info("Kafka consumer task started.")
    except Exception as exc:  # noqa: BLE001
        logger.warning("Kafka unavailable (%s) — running without consumer.", exc)

    yield

    if consume_task and not consume_task.done():
        consume_task.cancel()
        try:
            await consume_task
        except asyncio.CancelledError:
            pass

    await consumer.stop()
    logger.info("Shutdown complete.")


app = FastAPI(
    title="Analytics Service",
    description="Consumes analytics events from Kafka and provides query APIs for aggregated metrics.",
    version="1.0.0",
    lifespan=lifespan,
)

app.include_router(router)


if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=settings.HTTP_PORT,
        reload=False,
        log_level="info",
    )
