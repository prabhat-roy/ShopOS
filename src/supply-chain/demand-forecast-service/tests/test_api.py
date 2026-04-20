"""API-layer tests using HTTPX AsyncClient with mocked store."""

from __future__ import annotations

from datetime import datetime, timezone
from unittest.mock import AsyncMock, MagicMock

import pytest
import pytest_asyncio
from fastapi import FastAPI
from httpx import AsyncClient, ASGITransport

from forecast.api import router
from forecast.engine import ForecastEngine
from forecast.models import ForecastResponse
from forecast.store import AsyncPgStore


def _make_app(mock_store: AsyncPgStore) -> FastAPI:
    app = FastAPI()
    app.include_router(router)
    app.state.store = mock_store
    app.state.engine = ForecastEngine()
    return app


def _sample_forecast(product_id: str = "prod-test") -> ForecastResponse:
    return ForecastResponse(
        productId=product_id,
        forecastedDemand=210.0,
        averageDailySales=7.0,
        historicalDays=30,
        confidence=0.95,
        generatedAt=datetime(2024, 6, 1, 0, 0, 0, tzinfo=timezone.utc),
    )


# ------------------------------------------------------------------
# 1. GET /healthz returns 200 with status ok
# ------------------------------------------------------------------
@pytest.mark.asyncio
async def test_healthz() -> None:
    mock_store = MagicMock(spec=AsyncPgStore)
    app = _make_app(mock_store)
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as client:
        resp = await client.get("/healthz")
    assert resp.status_code == 200
    assert resp.json() == {"status": "ok"}


# ------------------------------------------------------------------
# 2. POST /forecasts/generate returns forecast for known product
# ------------------------------------------------------------------
@pytest.mark.asyncio
async def test_generate_forecast_success() -> None:
    mock_store = MagicMock(spec=AsyncPgStore)
    from datetime import date
    from forecast.models import SalesRecord

    sales = [
        SalesRecord(productId="prod-1", sku="SKU-1", quantity=5,
                    saleDate=date(2024, 5, i + 1), orderId=f"ord-{i}")
        for i in range(30)
    ]
    mock_store.get_sales_history = AsyncMock(return_value=sales)
    mock_store.save_forecast = AsyncMock()

    app = _make_app(mock_store)
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as client:
        resp = await client.post(
            "/forecasts/generate",
            json={"productId": "prod-1", "days": 30},
        )

    assert resp.status_code == 200
    body = resp.json()
    assert body["productId"] == "prod-1"
    assert body["forecastedDemand"] == pytest.approx(150.0, rel=1e-2)
    assert body["confidence"] == 0.95
    mock_store.save_forecast.assert_awaited_once()


# ------------------------------------------------------------------
# 3. GET /forecasts/{product_id} returns stored forecast
# ------------------------------------------------------------------
@pytest.mark.asyncio
async def test_get_forecast_found() -> None:
    mock_store = MagicMock(spec=AsyncPgStore)
    mock_store.get_forecast = AsyncMock(return_value=_sample_forecast("prod-abc"))

    app = _make_app(mock_store)
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as client:
        resp = await client.get("/forecasts/prod-abc")

    assert resp.status_code == 200
    assert resp.json()["productId"] == "prod-abc"


# ------------------------------------------------------------------
# 4. GET /forecasts/{product_id} returns 404 when not found
# ------------------------------------------------------------------
@pytest.mark.asyncio
async def test_get_forecast_not_found() -> None:
    mock_store = MagicMock(spec=AsyncPgStore)
    mock_store.get_forecast = AsyncMock(return_value=None)

    app = _make_app(mock_store)
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as client:
        resp = await client.get("/forecasts/nonexistent-product")

    assert resp.status_code == 404


# ------------------------------------------------------------------
# 5. GET /forecasts returns list of forecasts
# ------------------------------------------------------------------
@pytest.mark.asyncio
async def test_list_forecasts() -> None:
    mock_store = MagicMock(spec=AsyncPgStore)
    mock_store.list_forecasts = AsyncMock(
        return_value=[_sample_forecast("prod-1"), _sample_forecast("prod-2")]
    )

    app = _make_app(mock_store)
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as client:
        resp = await client.get("/forecasts")

    assert resp.status_code == 200
    body = resp.json()
    assert isinstance(body, list)
    assert len(body) == 2


# ------------------------------------------------------------------
# 6. POST /forecasts/alert returns alert when stock is low
# ------------------------------------------------------------------
@pytest.mark.asyncio
async def test_alert_returns_when_stock_low() -> None:
    mock_store = MagicMock(spec=AsyncPgStore)
    mock_store.get_forecast = AsyncMock(return_value=_sample_forecast("prod-low"))

    app = _make_app(mock_store)
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as client:
        resp = await client.post(
            "/forecasts/alert",
            json={"productId": "prod-low", "sku": "SKU-LOW", "currentStock": 50},
        )

    assert resp.status_code == 200
    body = resp.json()
    # forecastedDemand=210, threshold=420; currentStock=50 < 420 → alert triggered
    assert body is not None
    assert body["productId"] == "prod-low"
    assert body["currentStock"] == 50
