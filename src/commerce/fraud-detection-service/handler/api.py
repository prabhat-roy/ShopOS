"""FastAPI route handlers for the fraud-detection-service."""

from __future__ import annotations

from typing import Optional

from fastapi import APIRouter, Depends, HTTPException, Query, Request

from fraud.models import FraudCheckRequest, FraudCheckResult
from fraud.scorer import FraudScorer
from fraud.store import FraudStore

router = APIRouter()
_scorer = FraudScorer()


def _get_store(request: Request) -> FraudStore:
    return request.app.state.store


@router.get("/healthz")
async def healthz() -> dict:
    return {"status": "ok"}


@router.post("/fraud/check", response_model=FraudCheckResult, status_code=200)
async def check_fraud(
    req: FraudCheckRequest,
    store: FraudStore = Depends(_get_store),
) -> FraudCheckResult:
    result = _scorer.score(req)
    await store.save_result(result)
    return result


@router.get(
    "/fraud/results/{order_id}",
    response_model=FraudCheckResult,
    status_code=200,
)
async def get_result(
    order_id: str,
    store: FraudStore = Depends(_get_store),
) -> FraudCheckResult:
    result = await store.get_result(order_id)
    if result is None:
        raise HTTPException(status_code=404, detail=f"No fraud result for order '{order_id}'")
    return result


@router.get(
    "/fraud/flagged",
    response_model=list[FraudCheckResult],
    status_code=200,
)
async def list_flagged(
    limit: int = Query(default=50, ge=1, le=500),
    store: FraudStore = Depends(_get_store),
) -> list[FraudCheckResult]:
    return await store.list_flagged(limit=limit)
