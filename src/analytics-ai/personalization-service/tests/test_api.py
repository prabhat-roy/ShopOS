"""API tests for personalization-service — 8 tests with mocked service."""
from __future__ import annotations

from datetime import datetime, timezone
from typing import AsyncIterator
from unittest.mock import AsyncMock, MagicMock

import pytest
import pytest_asyncio
from httpx import ASGITransport, AsyncClient

from personalize.models import PersonalizedResult, UserProfile
from personalize.service import PersonalizationService
from main import app


def _sample_profile(user_id: str = "u-1") -> UserProfile:
    return UserProfile(
        userId=user_id,
        preferredCategories=["electronics"],
        preferredBrands=["Sony"],
        priceRangeLow=50.0,
        priceRangeHigh=5000.0,
        recentlyViewedProducts=["prod-10"],
        purchaseHistory=["prod-20"],
        excludedCategories=[],
        updatedAt=datetime.now(timezone.utc),
    )


def _sample_result(user_id: str = "u-1") -> PersonalizedResult:
    return PersonalizedResult(
        userId=user_id,
        rankedProductIds=["prod-1", "prod-2"],
        scores={"prod-1": 0.9, "prod-2": 0.4},
        contextType="homepage",
        profile=_sample_profile(user_id),
    )


def _mock_service() -> AsyncMock:
    svc = AsyncMock(spec=PersonalizationService)
    svc.personalize = AsyncMock(return_value=_sample_result())
    svc.get_profile = AsyncMock(return_value=_sample_profile())
    svc.update_profile = AsyncMock(return_value=_sample_profile())
    svc.record_view = AsyncMock(return_value=None)
    svc.record_purchase = AsyncMock(return_value=None)
    svc.delete_profile = AsyncMock(return_value=True)
    return svc


@pytest_asyncio.fixture
async def client() -> AsyncIterator[AsyncClient]:
    app.state.service = _mock_service()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        yield ac


@pytest.mark.asyncio
async def test_healthz(client: AsyncClient):
    response = await client.get("/healthz")
    assert response.status_code == 200
    assert response.json() == {"status": "ok"}


@pytest.mark.asyncio
async def test_personalize(client: AsyncClient):
    payload = {
        "userId": "u-1",
        "contextType": "homepage",
        "candidateProductIds": ["prod-1", "prod-2", "prod-3"],
        "limit": 10,
    }
    response = await client.post("/personalize", json=payload)
    assert response.status_code == 200
    data = response.json()
    assert data["userId"] == "u-1"
    assert isinstance(data["rankedProductIds"], list)
    assert isinstance(data["scores"], dict)


@pytest.mark.asyncio
async def test_get_profile(client: AsyncClient):
    response = await client.get("/profiles/u-1")
    assert response.status_code == 200
    data = response.json()
    assert data["userId"] == "u-1"
    assert "preferredCategories" in data


@pytest.mark.asyncio
async def test_get_profile_not_found(client: AsyncClient):
    app.state.service.get_profile = AsyncMock(return_value=None)
    response = await client.get("/profiles/nonexistent")
    assert response.status_code == 404


@pytest.mark.asyncio
async def test_upsert_profile(client: AsyncClient):
    payload = {
        "preferredCategories": ["books"],
        "preferredBrands": ["Penguin"],
        "priceRangeLow": 10.0,
        "priceRangeHigh": 100.0,
    }
    response = await client.put("/profiles/u-1", json=payload)
    assert response.status_code == 200
    data = response.json()
    assert data["userId"] == "u-1"


@pytest.mark.asyncio
async def test_record_view(client: AsyncClient):
    response = await client.post("/profiles/u-1/view", json={"productId": "prod-99"})
    assert response.status_code == 204


@pytest.mark.asyncio
async def test_record_purchase(client: AsyncClient):
    response = await client.post("/profiles/u-1/purchase", json={"productId": "prod-55"})
    assert response.status_code == 204


@pytest.mark.asyncio
async def test_delete_profile(client: AsyncClient):
    response = await client.delete("/profiles/u-1")
    assert response.status_code == 204
