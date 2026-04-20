"""email-service — entry point.

Starts the Kafka consumer as a background asyncio task, then launches
FastAPI via uvicorn on HTTP_PORT (default 8503).

The consumer is started/stopped using FastAPI lifespan hooks so that it
participates in the same event loop as the API handlers.
"""
from __future__ import annotations

import logging
import sys
from contextlib import asynccontextmanager
from typing import AsyncGenerator

import uvicorn
from fastapi import FastAPI

from app.api import router
from app.config import settings
from app.consumer import kafka_consumer

# ---------------------------------------------------------------------------
# Logging
# ---------------------------------------------------------------------------

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s  %(levelname)-8s  %(name)s  %(message)s",
    stream=sys.stdout,
)
logger = logging.getLogger("email-service")


# ---------------------------------------------------------------------------
# Lifespan
# ---------------------------------------------------------------------------


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncGenerator[None, None]:
    """Manage the Kafka consumer lifecycle alongside the FastAPI app."""
    logger.info("Starting email-service — port=%d", settings.HTTP_PORT)
    try:
        await kafka_consumer.start()
        logger.info("Kafka consumer started successfully")
    except Exception as exc:
        # Log but do NOT abort startup — the health endpoint still needs to be
        # reachable so that Kubernetes can report the correct state.
        logger.error("Failed to start Kafka consumer: %s", exc)

    yield  # Application runs here

    logger.info("Shutting down email-service")
    await kafka_consumer.stop()


# ---------------------------------------------------------------------------
# Application
# ---------------------------------------------------------------------------


def create_app() -> FastAPI:
    application = FastAPI(
        title="email-service",
        description="ShopOS communications — simulated email delivery via Kafka",
        version="1.0.0",
        lifespan=lifespan,
    )
    application.include_router(router)
    return application


app = create_app()


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------


if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=settings.HTTP_PORT,
        log_level="info",
        access_log=True,
    )
