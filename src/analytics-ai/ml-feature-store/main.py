import logging
from contextlib import asynccontextmanager

import uvicorn
from fastapi import FastAPI

from features.api import router
from features.config import settings
from features.store import AsyncPgStore

logging.basicConfig(level=settings.LOG_LEVEL.upper())
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    store = AsyncPgStore(settings.DATABASE_URL)
    try:
        await store.connect()
        app.state.store = store
    except Exception as exc:
        logger.warning("PostgreSQL unavailable at startup: %s — running without persistence", exc)
        app.state.store = None
    logger.info(
        "ml-feature-store started on HTTP :%d gRPC :%d",
        settings.HTTP_PORT,
        settings.GRPC_PORT,
    )
    yield
    try:
        await store.disconnect()
    except Exception:
        pass
    logger.info("ml-feature-store stopped")


app = FastAPI(
    title="ML Feature Store",
    description="Stores and serves ML features for model training and inference.",
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
        log_level=settings.LOG_LEVEL.lower(),
    )
