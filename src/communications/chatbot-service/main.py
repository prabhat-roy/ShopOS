"""
chatbot-service — entry point.

HTTP  : port 8193 (FastAPI / uvicorn)
gRPC  : port 50189
Redis : conversation state storage
"""

from __future__ import annotations

import logging
import sys

import uvicorn
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from chatbot.api import router
from chatbot.config import settings

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
    title="Chatbot Service",
    description=(
        "Rule-based + intent-classification chatbot for Tier-1 customer support deflection. "
        "Handles order status, returns, and FAQ queries. Stores conversation state in Redis."
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
# Startup / shutdown
# ---------------------------------------------------------------------------


@app.on_event("startup")
async def startup() -> None:
    try:
        import redis.asyncio as aioredis  # type: ignore[import]

        pool = aioredis.ConnectionPool.from_url(
            settings.REDIS_URL,
            max_connections=20,
            decode_responses=True,
        )
        app.state.redis = aioredis.Redis(connection_pool=pool)
        await app.state.redis.ping()
        logger.info("Redis connection ready")
    except Exception as exc:
        logger.warning("Redis unavailable — running without session persistence: %s", exc)
        app.state.redis = None

    logger.info("chatbot-service starting on HTTP :%d", settings.HTTP_PORT)


@app.on_event("shutdown")
async def shutdown() -> None:
    if getattr(app.state, "redis", None) is not None:
        await app.state.redis.close()
        logger.info("Redis connection closed")

    logger.info("chatbot-service shut down")


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
