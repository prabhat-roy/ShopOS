import logging
from typing import Annotated

from fastapi import APIRouter, Depends, HTTPException, Query, Request, status

from search.indexer import ProductIndexer
from search.models import ProductDocument, SearchRequest, SearchResult, SuggestRequest
from search.searcher import ProductSearcher

logger = logging.getLogger(__name__)

router = APIRouter()


# ---------------------------------------------------------------------------
# Dependency helpers — pull shared instances stored on app.state
# ---------------------------------------------------------------------------


def get_searcher(request: Request) -> ProductSearcher:
    return request.app.state.searcher


def get_indexer(request: Request) -> ProductIndexer:
    return request.app.state.indexer


SearcherDep = Annotated[ProductSearcher, Depends(get_searcher)]
IndexerDep = Annotated[ProductIndexer, Depends(get_indexer)]


# ---------------------------------------------------------------------------
# Routes
# ---------------------------------------------------------------------------


@router.get("/healthz", tags=["health"])
async def healthz() -> dict[str, str]:
    """Liveness probe — returns 200 {"status": "ok"}."""
    return {"status": "ok"}


@router.post(
    "/search/products",
    response_model=SearchResult,
    status_code=status.HTTP_200_OK,
    tags=["search"],
)
async def search_products(
    req: SearchRequest,
    searcher: SearcherDep,
) -> SearchResult:
    """
    Full-text product search with faceted filtering.

    - Fuzzy multi-match across name, description, sku, and tags.
    - Keyword filters: category_id, brand_id, status, tags.
    - Price range filter via min_price / max_price.
    - Aggregations: by_category, by_brand, price_range, price_stats.
    """
    try:
        return await searcher.search(req)
    except Exception as exc:
        logger.exception("Search failed: %s", exc)
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Search request failed",
        ) from exc


@router.get(
    "/search/suggest",
    response_model=list[str],
    status_code=status.HTTP_200_OK,
    tags=["search"],
)
async def autocomplete(
    prefix: Annotated[str, Query(min_length=1, description="Prefix string for autocomplete")],
    size: Annotated[int, Query(ge=1, le=20, description="Max suggestions to return")] = 5,
    searcher: SearcherDep = ...,  # type: ignore[assignment]
) -> list[str]:
    """Return up to *size* product-name suggestions for *prefix*."""
    try:
        return await searcher.suggest(prefix=prefix, size=size)
    except Exception as exc:
        logger.exception("Suggest failed: %s", exc)
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Suggest request failed",
        ) from exc


@router.post(
    "/search/index",
    status_code=status.HTTP_201_CREATED,
    tags=["index"],
)
async def index_product(
    doc: ProductDocument,
    indexer: IndexerDep,
) -> dict[str, str]:
    """Index (or re-index) a single product document."""
    try:
        await indexer.index_product(doc)
        return {"id": doc.id, "result": "indexed"}
    except Exception as exc:
        logger.exception("Index failed for product %s: %s", doc.id, exc)
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Indexing failed",
        ) from exc


@router.delete(
    "/search/index/{product_id}",
    status_code=status.HTTP_204_NO_CONTENT,
    tags=["index"],
)
async def delete_from_index(
    product_id: str,
    indexer: IndexerDep,
) -> None:
    """Remove a product from the search index."""
    try:
        await indexer.delete_product(product_id)
    except Exception as exc:
        logger.exception("Delete failed for product %s: %s", product_id, exc)
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Delete from index failed",
        ) from exc


@router.post(
    "/search/bulk-index",
    status_code=status.HTTP_200_OK,
    tags=["index"],
)
async def bulk_index_products(
    docs: list[ProductDocument],
    indexer: IndexerDep,
) -> dict[str, int]:
    """Bulk-index a list of product documents."""
    if not docs:
        return {"indexed": 0, "errors": 0}
    try:
        success, errors = await indexer.bulk_index(docs)
        return {"indexed": success, "errors": len(errors)}
    except Exception as exc:
        logger.exception("Bulk index failed: %s", exc)
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Bulk indexing failed",
        ) from exc
