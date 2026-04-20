from __future__ import annotations

from typing import Annotated

from fastapi import APIRouter, Depends, HTTPException, Request, status

from .engine import PriceOptimizer
from .models import PriceOptimizationRequest, PriceOptimizationResult, PricingRule
from .store import AsyncPgStore

router = APIRouter()
_optimizer = PriceOptimizer()


def _get_store(request: Request) -> AsyncPgStore:
    return request.app.state.store


# ---------------------------------------------------------------------------
# Health
# ---------------------------------------------------------------------------


@router.get("/healthz", tags=["health"])
async def healthz() -> dict:
    return {"status": "ok"}


# ---------------------------------------------------------------------------
# Optimize
# ---------------------------------------------------------------------------


@router.post(
    "/optimize",
    response_model=PriceOptimizationResult,
    status_code=status.HTTP_200_OK,
    tags=["optimize"],
)
async def optimize(
    req: PriceOptimizationRequest,
    store: Annotated[AsyncPgStore, Depends(_get_store)],
) -> PriceOptimizationResult:
    try:
        result = _optimizer.optimize(req)
    except ValueError as exc:
        raise HTTPException(status_code=status.HTTP_422_UNPROCESSABLE_ENTITY, detail=str(exc))
    await store.save_result(result)
    return result


@router.post(
    "/optimize/batch",
    response_model=list[PriceOptimizationResult],
    status_code=status.HTTP_200_OK,
    tags=["optimize"],
)
async def optimize_batch(
    requests: list[PriceOptimizationRequest],
    store: Annotated[AsyncPgStore, Depends(_get_store)],
) -> list[PriceOptimizationResult]:
    results: list[PriceOptimizationResult] = []
    for req in requests:
        try:
            result = _optimizer.optimize(req)
        except ValueError as exc:
            raise HTTPException(
                status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
                detail=f"Error for product {req.productId}: {exc}",
            )
        await store.save_result(result)
        results.append(result)
    return results


# ---------------------------------------------------------------------------
# Optimization history
# ---------------------------------------------------------------------------


@router.get(
    "/optimization/{product_id}",
    response_model=PriceOptimizationResult,
    tags=["optimize"],
)
async def get_latest_optimization(
    product_id: str,
    store: Annotated[AsyncPgStore, Depends(_get_store)],
) -> PriceOptimizationResult:
    result = await store.get_result(product_id)
    if result is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"No optimization result found for product '{product_id}'",
        )
    return result


# ---------------------------------------------------------------------------
# Rules
# ---------------------------------------------------------------------------


@router.get(
    "/rules",
    response_model=list[PricingRule],
    tags=["rules"],
)
async def list_rules(
    store: Annotated[AsyncPgStore, Depends(_get_store)],
) -> list[PricingRule]:
    return await store.list_rules()


@router.post(
    "/rules",
    response_model=PricingRule,
    status_code=status.HTTP_200_OK,
    tags=["rules"],
)
async def upsert_rule(
    rule: PricingRule,
    store: Annotated[AsyncPgStore, Depends(_get_store)],
) -> PricingRule:
    if rule.minPrice > rule.maxPrice:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail="minPrice must be <= maxPrice",
        )
    await store.save_rule(rule)
    return rule


@router.delete(
    "/rules/{product_id}",
    status_code=status.HTTP_204_NO_CONTENT,
    response_model=None,
    tags=["rules"],
)
async def delete_rule(
    product_id: str,
    store: Annotated[AsyncPgStore, Depends(_get_store)],
) -> None:
    deleted = await store.delete_rule(product_id)
    if not deleted:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Rule for product '{product_id}' not found",
        )
