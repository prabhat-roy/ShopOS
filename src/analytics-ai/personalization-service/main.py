import logging
from contextlib import asynccontextmanager

import uvicorn
from fastapi import FastAPI

from personalize.api import router
from personalize.config import settings
from personalize.repository import MongoRepository
from personalize.service import PersonalizationService

logging.basicConfig(level=settings.LOG_LEVEL.upper())
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    repo = MongoRepository(settings.MONGODB_URI, settings.MONGODB_DB)
    try:
        await repo.connect()
        app.state.service = PersonalizationService(repo)
    except Exception as exc:
        logger.warning("MongoDB unavailable at startup: %s — running without persistence", exc)
        app.state.service = None
    logger.info(
        "personalization-service started on HTTP :%d gRPC :%d",
        settings.HTTP_PORT,
        settings.GRPC_PORT,
    )
    yield
    try:
        await repo.disconnect()
    except Exception:
        pass
    logger.info("personalization-service stopped")


app = FastAPI(
    title="Personalization Service",
    description="Personalizes product listings and content for users based on preferences and behavior.",
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
