"""Entry point for the fraud-detection-service."""

from contextlib import asynccontextmanager

import uvicorn
from fastapi import FastAPI

from fraud.config import settings
from fraud.store import FraudStore
from handler.api import router


import logging

logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    try:
        app.state.store = await FraudStore.create(settings.database_url)
    except Exception as exc:
        logger.warning("PostgreSQL unavailable — running without persistence: %s", exc)
        app.state.store = None
    yield
    if app.state.store is not None:
        await app.state.store.close()


app = FastAPI(
    title="Fraud Detection Service",
    description="Rule-based fraud scoring and result persistence for ShopOS commerce domain.",
    version="1.0.0",
    lifespan=lifespan,
)

app.include_router(router)

if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=settings.http_port,
        reload=False,
    )
