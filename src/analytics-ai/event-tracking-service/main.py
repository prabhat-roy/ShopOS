from __future__ import annotations

import asyncio
import logging
from contextlib import asynccontextmanager
from typing import AsyncIterator

import uvicorn
from fastapi import FastAPI

from tracker.api import router
from tracker.config import settings
from tracker.publisher import KafkaEventPublisher
from tracker.service import EventTrackingService
from tracker.store import CassandraEventStore, EventStore, InMemoryEventStore

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(levelname)s %(name)s — %(message)s",
)
logger = logging.getLogger(__name__)


def _build_store() -> EventStore:
    if settings.STORE_BACKEND == "memory":
        logger.info("Using InMemoryEventStore (testing mode)")
        return InMemoryEventStore()

    try:
        from cassandra.cluster import Cluster  # type: ignore[import-untyped]

        cluster = Cluster(settings.cassandra_host_list, connect_timeout=2)
        session = cluster.connect()
        logger.info("Connected to Cassandra at %s", settings.cassandra_host_list)
        return CassandraEventStore(session, settings.CASSANDRA_KEYSPACE)
    except Exception as exc:  # noqa: BLE001
        logger.warning(
            "Cassandra unavailable (%s) — falling back to InMemoryEventStore", exc
        )
        return InMemoryEventStore()


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncIterator[None]:
    store = _build_store()
    publisher = KafkaEventPublisher()

    try:
        await publisher.start()
    except Exception as exc:  # noqa: BLE001
        logger.warning("Kafka unavailable (%s) — publishing will be skipped.", exc)

    service = EventTrackingService(store=store, publisher=publisher)
    app.state.service = service
    app.state.store = store
    app.state.publisher = publisher

    logger.info("EventTrackingService ready on port %d", settings.HTTP_PORT)

    yield

    await publisher.stop()
    logger.info("Shutdown complete.")


app = FastAPI(
    title="Event Tracking Service",
    description="Receives client-side events via HTTP and publishes to Kafka for downstream processing.",
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
