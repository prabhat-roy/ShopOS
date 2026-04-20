from __future__ import annotations

from collections import defaultdict
from typing import Dict, List, Optional

from fastapi import APIRouter
from pydantic import BaseModel

router = APIRouter()

# In-memory store for demo/testing; real impl uses Cassandra
_search_events: List[dict] = []


class SearchEvent(BaseModel):
    query: str
    result_count: int
    customer_id: Optional[str] = None
    session_id: Optional[str] = None
    clicked_product_id: Optional[str] = None
    converted: bool = False
    timestamp: Optional[str] = None


class SearchStatsResponse(BaseModel):
    total_searches: int
    zero_result_searches: int
    zero_result_rate: float
    top_queries: List[dict]
    avg_results_per_search: float


@router.get("/healthz")
async def health():
    return {"status": "ok"}


@router.post("/search-events", status_code=201)
async def ingest_event(event: SearchEvent):
    """Ingest a search event (HTTP path — Kafka path preferred in production)."""
    _search_events.append(event.model_dump())
    return {"status": "accepted"}


@router.get("/search-analytics/stats", response_model=SearchStatsResponse)
async def get_stats():
    """
    Return aggregate search quality metrics:
    total searches, zero-result rate, top queries, average results per search.
    """
    total = len(_search_events)
    if total == 0:
        return SearchStatsResponse(
            total_searches=0,
            zero_result_searches=0,
            zero_result_rate=0.0,
            top_queries=[],
            avg_results_per_search=0.0,
        )

    zero_result = sum(1 for e in _search_events if e["result_count"] == 0)
    query_counts: Dict[str, int] = defaultdict(int)
    for e in _search_events:
        query_counts[e["query"]] += 1

    top_queries = sorted(
        [{"query": q, "count": c} for q, c in query_counts.items()],
        key=lambda x: x["count"],
        reverse=True,
    )[:20]

    avg_results = sum(e["result_count"] for e in _search_events) / total

    return SearchStatsResponse(
        total_searches=total,
        zero_result_searches=zero_result,
        zero_result_rate=round(zero_result / total, 4),
        top_queries=top_queries,
        avg_results_per_search=round(avg_results, 2),
    )


@router.get("/search-analytics/zero-results")
async def get_zero_result_queries():
    """Return queries that produced zero results — used for search quality tuning."""
    zero = [e["query"] for e in _search_events if e["result_count"] == 0]
    counts: Dict[str, int] = defaultdict(int)
    for q in zero:
        counts[q] += 1
    return {
        "zero_result_queries": sorted(
            [{"query": q, "count": c} for q, c in counts.items()],
            key=lambda x: x["count"],
            reverse=True,
        )
    }
