"""API tests for price-optimization-service — 6 tests."""
from __future__ import annotations

from datetime import datetime, timezone
from typing import AsyncIterator
from unittest.mock import AsyncMock, MagicMock

import pytest
import pytest_asyncio
from httpx import ASGITransport, AsyncClient

from main import app
from optimizer.models import PriceOptimizationResult, PricingRule
from optimizer.store import AsyncPgStore


def _mock_store() -> AsyncMock:
    store = AsyncMock(spec=AsyncPgStore)
    return store


def _sample_result(product_id: str = "prod-001") -> PriceOptimizationResult:
    return PriceOptimizationResult(
        productId=product_id,
        currentPrice=100.0,
        suggestedPrice=95.0,
        expectedDemandChange=0.075,
        expectedRevenueChange=-0.023,
        marginAtSuggestedPrice=0.45,
        confidence=0.82,
        reasoning="Price reduced to stimulate demand.",
        generatedAt=datetime.now(timezone.utc),
    )


@pytest_asyncio.fixture
async def client() -> AsyncIterator[AsyncClient]:
    store = _mock_store()
    store.save_result = AsyncMock(return_value=None)
    store.get_result = AsyncMock(return_value=_sample_result())
    store.list_rules = AsyncMock(return_value=[])
    store.save_rule = AsyncMock(return_value=None)
    store.delete_rule = AsyncMock(return_value=True)

    app.state.store = store
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        yield ac


@pytest.mark.asyncio
async def test_healthz(client: AsyncClient):
    response = await client.get("/healthz")
    assert response.status_code == 200
    assert response.json() == {"status": "ok"}


@pytest.mark.asyncio
async def test_optimize_returns_result(client: AsyncClient):
    payload = {
        "productId": "prod-001",
        "currentPrice": 100.0,
        "costPrice": 50.0,
        "minPrice": 60.0,
        "maxPrice": 150.0,
        "targetMargin": 0.3,
        "elasticity": -1.5,
    }
    response = await client.post("/optimize", json=payload)
    assert response.status_code == 200
    data = response.json()
    assert data["productId"] == "prod-001"
    assert "suggestedPrice" in data
    assert "confidence" in data


@pytest.mark.asyncio
async def test_optimize_batch(client: AsyncClient):
    payloads = [
        {
            "productId": f"prod-00{i}",
            "currentPrice": 100.0,
            "costPrice": 40.0,
            "minPrice": 50.0,
            "maxPrice": 160.0,
            "targetMargin": 0.25,
            "elasticity": -1.2,
        }
        for i in range(1, 4)
    ]
    response = await client.post("/optimize/batch", json=payloads)
    assert response.status_code == 200
    results = response.json()
    assert len(results) == 3
    product_ids = {r["productId"] for r in results}
    assert product_ids == {"prod-001", "prod-002", "prod-003"}


@pytest.mark.asyncio
async def test_get_latest_optimization(client: AsyncClient):
    response = await client.get("/optimization/prod-001")
    assert response.status_code == 200
    data = response.json()
    assert data["productId"] == "prod-001"


@pytest.mark.asyncio
async def test_get_optimization_not_found(client: AsyncClient):
    app.state.store.get_result = AsyncMock(return_value=None)
    response = await client.get("/optimization/nonexistent")
    assert response.status_code == 404


@pytest.mark.asyncio
async def test_upsert_rule(client: AsyncClient):
    payload = {
        "productId": "prod-001",
        "minPrice": 60.0,
        "maxPrice": 140.0,
        "targetMargin": 0.35,
        "active": True,
    }
    response = await client.post("/rules", json=payload)
    assert response.status_code == 200
    data = response.json()
    assert data["productId"] == "prod-001"
    assert data["active"] is True
