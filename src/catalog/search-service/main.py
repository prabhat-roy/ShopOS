import logging
from contextlib import asynccontextmanager
from typing import AsyncGenerator

import uvicorn
from fastapi import FastAPI

from handler.api import router
from search.config import settings
from search.indexer import ProductIndexer
from search.searcher import ProductSearcher

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(levelname)s %(name)s — %(message)s",
)
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncGenerator[None, None]:
    """Startup: create ES index and shared clients. Shutdown: close connections."""
    logger.info("Starting search-service on port %d", settings.http_port)

    indexer = ProductIndexer()
    searcher = ProductSearcher(client=indexer.client)  # share one ES client

    try:
        await indexer.create_index_if_not_exists()
    except Exception as exc:  # noqa: BLE001
        logger.warning("Could not create Elasticsearch index at startup: %s", exc)

    app.state.indexer = indexer
    app.state.searcher = searcher

    logger.info("search-service ready")
    yield

    logger.info("Shutting down search-service")
    await indexer.close()


app = FastAPI(
    title="search-service",
    description=(
        "Full-text product search with faceted filtering. "
        "Indexes products from catalog events; supports fuzzy search, "
        "autocomplete, and aggregations."
    ),
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
        access_log=True,
    )
