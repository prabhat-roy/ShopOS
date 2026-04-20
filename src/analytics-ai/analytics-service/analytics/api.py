from __future__ import annotations

import logging
from datetime import datetime
from typing import List, Optional

from fastapi import APIRouter, Depends, HTTPException, Query, Request

from analytics.models import (
    AnalyticsSummary,
    PageViewRecord,
    SearchQueryStat,
    TopProduct,
)
from analytics.store import AnalyticsStore

logger = logging.getLogger(__name__)
router = APIRouter()


def _get_store(request: Request) -> AnalyticsStore:
    return request.app.state.store  # type: ignore[no-any-return]


def _parse_date(value: str, field_name: str) -> datetime:
    for fmt in ("%Y-%m-%dT%H:%M:%S", "%Y-%m-%dT%H:%M:%SZ", "%Y-%m-%d"):
        try:
            return datetime.strptime(value, fmt)
        except ValueError:
            continue
    raise HTTPException(
        status_code=422,
        detail=f"Invalid date format for '{field_name}'. Use YYYY-MM-DD or ISO-8601.",
    )


@router.get("/healthz")
async def healthz() -> dict:
    return {"status": "ok", "service": "analytics-service"}


@router.get("/analytics/page-views", response_model=List[PageViewRecord])
async def get_page_views(
    start: str = Query(..., description="Start date (YYYY-MM-DD or ISO-8601)"),
    end: str = Query(..., description="End date (YYYY-MM-DD or ISO-8601)"),
    limit: int = Query(default=100, ge=1, le=1000),
    store: AnalyticsStore = Depends(_get_store),
) -> List[PageViewRecord]:
    start_dt = _parse_date(start, "start")
    end_dt = _parse_date(end, "end")
    if end_dt < start_dt:
        raise HTTPException(status_code=422, detail="'end' must be >= 'start'.")
    return await store.get_page_views(start_dt, end_dt, limit)


@router.get("/analytics/top-products", response_model=List[TopProduct])
async def get_top_products(
    start: str = Query(..., description="Start date (YYYY-MM-DD or ISO-8601)"),
    end: str = Query(..., description="End date (YYYY-MM-DD or ISO-8601)"),
    limit: int = Query(default=100, ge=1, le=1000),
    store: AnalyticsStore = Depends(_get_store),
) -> List[TopProduct]:
    start_dt = _parse_date(start, "start")
    end_dt = _parse_date(end, "end")
    if end_dt < start_dt:
        raise HTTPException(status_code=422, detail="'end' must be >= 'start'.")
    return await store.get_top_products(start_dt, end_dt, limit)


@router.get("/analytics/searches", response_model=List[SearchQueryStat])
async def get_searches(
    start: str = Query(..., description="Start date (YYYY-MM-DD or ISO-8601)"),
    end: str = Query(..., description="End date (YYYY-MM-DD or ISO-8601)"),
    limit: int = Query(default=100, ge=1, le=1000),
    store: AnalyticsStore = Depends(_get_store),
) -> List[SearchQueryStat]:
    start_dt = _parse_date(start, "start")
    end_dt = _parse_date(end, "end")
    if end_dt < start_dt:
        raise HTTPException(status_code=422, detail="'end' must be >= 'start'.")
    return await store.get_search_queries(start_dt, end_dt, limit)


@router.get("/analytics/summary", response_model=AnalyticsSummary)
async def get_summary(
    start: str = Query(..., description="Start date (YYYY-MM-DD or ISO-8601)"),
    end: str = Query(..., description="End date (YYYY-MM-DD or ISO-8601)"),
    store: AnalyticsStore = Depends(_get_store),
) -> AnalyticsSummary:
    start_dt = _parse_date(start, "start")
    end_dt = _parse_date(end, "end")
    if end_dt < start_dt:
        raise HTTPException(status_code=422, detail="'end' must be >= 'start'.")
    return await store.get_summary(start_dt, end_dt)
