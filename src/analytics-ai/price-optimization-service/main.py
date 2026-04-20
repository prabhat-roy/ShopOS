import logging
from contextlib import asynccontextmanager

import uvicorn
from fastapi import FastAPI

from optimizer.api import router
from optimizer.config import settings
from optimizer.store import AsyncPgStore

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
        "price-optimization-service started on HTTP :%d gRPC :%d",
        settings.HTTP_PORT,
        settings.GRPC_PORT,
    )
    yield
    try:
        await store.disconnect()
    except Exception:
        pass
    logger.info("price-optimization-service stopped")


app = FastAPI(
    title="Price Optimization Service",
    description="Suggests optimized prices based on demand elasticity, competitor pricing, and margin targets.",
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
