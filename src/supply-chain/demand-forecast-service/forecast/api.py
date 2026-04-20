from __future__ import annotations

import logging
from typing import Optional

from fastapi import APIRouter, Depends, HTTPException, Request, status
from pydantic import BaseModel

from forecast.engine import ForecastEngine
from forecast.models import ForecastRequest, ForecastResponse, InventoryAlert
from forecast.store import AsyncPgStore

logger = logging.getLogger(__name__)
router = APIRouter()


# ---------------------------------------------------------------------------
# Dependency helpers
# ---------------------------------------------------------------------------

def _get_store(request: Request) -> AsyncPgStore:
    return request.app.state.store  # type: ignore[no-any-return]


def _get_engine(request: Request) -> ForecastEngine:
    return request.app.state.engine  # type: ignore[no-any-return]


# ---------------------------------------------------------------------------
# Request body for alert endpoint
# ---------------------------------------------------------------------------

class AlertRequest(BaseModel):
    productId: str
    sku: str = ""
    currentStock: int


# ---------------------------------------------------------------------------
# Routes
# ---------------------------------------------------------------------------

@router.get("/healthz", tags=["ops"])
async def healthz() -> dict:
    return {"status": "ok"}


@router.post(
    "/forecasts/generate",
    response_model=ForecastResponse,
    status_code=status.HTTP_200_OK,
    tags=["forecasts"],
)
async def generate_forecast(
    req: ForecastRequest,
    store: AsyncPgStore = Depends(_get_store),
    engine: ForecastEngine = Depends(_get_engine),
) -> ForecastResponse:
    """Query sales history, run the moving-average engine, persist and return the forecast."""
    history = await store.get_sales_history(req.productId, req.days)
    forecast = engine.compute_forecast(history, req.days)
    # Override productId in case history was empty
    forecast = forecast.model_copy(update={"productId": req.productId})
    await store.save_forecast(forecast)
    logger.info(
        "Forecast generated. productId=%s forecastedDemand=%.2f confidence=%.2f",
        req.productId,
        forecast.forecastedDemand,
        forecast.confidence,
    )
    return forecast


@router.get(
    "/forecasts/{product_id}",
    response_model=ForecastResponse,
    tags=["forecasts"],
)
async def get_forecast(
    product_id: str,
    store: AsyncPgStore = Depends(_get_store),
) -> ForecastResponse:
    forecast = await store.get_forecast(product_id)
    if forecast is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"No forecast found for product '{product_id}'",
        )
    return forecast


@router.get(
    "/forecasts",
    response_model=list[ForecastResponse],
    tags=["forecasts"],
)
async def list_forecasts(
    limit: int = 50,
    store: AsyncPgStore = Depends(_get_store),
) -> list[ForecastResponse]:
    return await store.list_forecasts(limit=min(limit, 200))


@router.post(
    "/forecasts/alert",
    response_model=Optional[InventoryAlert],
    tags=["forecasts"],
)
async def get_inventory_alert(
    req: AlertRequest,
    store: AsyncPgStore = Depends(_get_store),
    engine: ForecastEngine = Depends(_get_engine),
) -> Optional[InventoryAlert]:
    """
    Return an InventoryAlert if current stock is below 2x the latest stored forecast
    for the product, or None if stock levels are adequate.
    """
    forecast = await store.get_forecast(req.productId)
    if forecast is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"No forecast found for product '{req.productId}'. Generate a forecast first.",
        )
    alert = engine.generate_inventory_alert(
        product_id=req.productId,
        sku=req.sku,
        current_stock=req.currentStock,
        forecast=forecast,
    )
    return alert
