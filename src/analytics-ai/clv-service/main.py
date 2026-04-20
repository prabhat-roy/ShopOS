"""
clv-service — entry point.

HTTP  : port 8195 (FastAPI / uvicorn)
gRPC  : port 50190
DB    : PostgreSQL (CLV scores, customer segments)
"""

from __future__ import annotations

import logging
import sys

import uvicorn
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from clv.api import router
from clv.config import settings

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
    title="CLV Service",
    description=(
        "Customer Lifetime Value prediction and segmentation. "
        "Computes CLV scores using order history and behavioural features, "
        "and segments customers into tiers (Platinum, Gold, Silver, Bronze)."
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
    logger.info("clv-service starting on HTTP :%d | gRPC :%d", settings.HTTP_PORT, settings.GRPC_PORT)


@app.on_event("shutdown")
async def shutdown() -> None:
    logger.info("clv-service shut down")


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
