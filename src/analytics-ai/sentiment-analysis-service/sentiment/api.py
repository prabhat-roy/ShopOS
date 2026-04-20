from __future__ import annotations

import logging
from typing import Optional

from fastapi import APIRouter, HTTPException, Query, Request

from .analyzer import analyzer
from .models import (
    AggregateStats,
    AnalyzeRequest,
    BatchAnalyzeRequest,
    SentimentResult,
)

logger = logging.getLogger(__name__)

router = APIRouter()


def _get_store(request: Request):
    """Retrieve the AsyncPgStore attached to the application state."""
    return request.app.state.store


# ------------------------------------------------------------------
# Health
# ------------------------------------------------------------------


@router.get("/healthz", tags=["ops"])
async def healthz() -> dict:
    return {"status": "ok", "service": "sentiment-analysis-service"}


# ------------------------------------------------------------------
# Analysis
# ------------------------------------------------------------------


@router.post(
    "/sentiment/analyze",
    response_model=SentimentResult,
    tags=["sentiment"],
)
async def analyze_single(req: AnalyzeRequest, request: Request) -> SentimentResult:
    """Analyze sentiment for a single piece of text."""
    result = analyzer.analyze(
        text=req.text,
        entity_id=req.entityId,
        entity_type=req.entityType,
    )

    store = _get_store(request)
    if store is not None:
        try:
            await store.save_result(result)
        except Exception as exc:
            logger.warning("Failed to persist sentiment result: %s", exc)

    return result


@router.post(
    "/sentiment/batch",
    response_model=list[SentimentResult],
    tags=["sentiment"],
)
async def analyze_batch(req: BatchAnalyzeRequest, request: Request) -> list[SentimentResult]:
    """Analyze sentiment for up to 50 texts in a single request."""
    results = analyzer.batch_analyze(req.texts)

    store = _get_store(request)
    if store is not None:
        for result in results:
            try:
                await store.save_result(result)
            except Exception as exc:
                logger.warning("Failed to persist batch result: %s", exc)

    return results


# ------------------------------------------------------------------
# Retrieval
# ------------------------------------------------------------------


@router.get(
    "/sentiment/stats",
    response_model=AggregateStats,
    tags=["sentiment"],
)
async def get_stats(
    request: Request,
    entityType: Optional[str] = Query(default=None),
) -> AggregateStats:
    """Return aggregate sentiment statistics, optionally filtered by entity type."""
    store = _get_store(request)
    if store is None:
        raise HTTPException(status_code=503, detail="Database unavailable")

    stats = await store.get_aggregate_stats(entity_type=entityType)
    return stats


@router.get(
    "/sentiment",
    response_model=list[SentimentResult],
    tags=["sentiment"],
)
async def list_results(
    request: Request,
    entityType: Optional[str] = Query(default=None),
    limit: int = Query(default=50, ge=1, le=500),
) -> list[SentimentResult]:
    """List sentiment analysis results, optionally filtered by entity type."""
    store = _get_store(request)
    if store is None:
        raise HTTPException(status_code=503, detail="Database unavailable")

    results = await store.list_results(entity_type=entityType, limit=limit)
    return results


@router.get(
    "/sentiment/{entityId}",
    response_model=SentimentResult,
    tags=["sentiment"],
)
async def get_by_entity(entityId: str, request: Request) -> SentimentResult:
    """Retrieve the latest sentiment analysis result for an entity."""
    store = _get_store(request)
    if store is None:
        raise HTTPException(status_code=503, detail="Database unavailable")

    result = await store.get_result(entityId)
    if result is None:
        raise HTTPException(
            status_code=404,
            detail=f"No sentiment result found for entity '{entityId}'",
        )
    return result
