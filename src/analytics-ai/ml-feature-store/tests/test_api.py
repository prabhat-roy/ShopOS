"""API tests for ml-feature-store — 8 tests."""
from __future__ import annotations

import json
from datetime import datetime, timezone
from typing import AsyncIterator
from unittest.mock import AsyncMock

import pytest
import pytest_asyncio
from httpx import ASGITransport, AsyncClient

from features.models import FeatureDefinition, FeatureType, FeatureValue, FeatureVector
from features.store import AsyncPgStore
from main import app


def _mock_store() -> AsyncMock:
    return AsyncMock(spec=AsyncPgStore)


def _sample_definition() -> FeatureDefinition:
    return FeatureDefinition(
        name="age",
        featureGroup="user",
        type=FeatureType.NUMERIC,
        description="User age",
        tags=["demographic"],
        defaultValue=0,
    )


def _sample_vector() -> FeatureVector:
    return FeatureVector(
        entityId="user-1",
        features={"age": 25, "score": 0.9},
        missingFeatures=[],
        retrievedAt=datetime.now(timezone.utc),
    )


def _sample_fv() -> FeatureValue:
    return FeatureValue(
        entityId="user-1",
        featureName="age",
        featureGroup="user",
        value=25,
        version=1,
        computedAt=datetime.now(timezone.utc),
    )


@pytest_asyncio.fixture
async def client() -> AsyncIterator[AsyncClient]:
    store = _mock_store()
    store.register_feature = AsyncMock(return_value=_sample_definition())
    store.list_definitions = AsyncMock(return_value=[_sample_definition()])
    store.get_definition = AsyncMock(return_value=_sample_definition())
    store.save_value = AsyncMock(return_value=None)
    store.save_batch = AsyncMock(return_value=None)
    store.get_feature_vector = AsyncMock(return_value=_sample_vector())
    store.get_entity_features = AsyncMock(return_value=[_sample_fv()])
    store.delete_entity_features = AsyncMock(return_value=1)

    app.state.store = store
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        yield ac


@pytest.mark.asyncio
async def test_healthz(client: AsyncClient):
    response = await client.get("/healthz")
    assert response.status_code == 200
    assert response.json() == {"status": "ok"}


@pytest.mark.asyncio
async def test_register_feature(client: AsyncClient):
    payload = {
        "name": "age",
        "featureGroup": "user",
        "type": "NUMERIC",
        "description": "User age",
        "tags": ["demographic"],
        "defaultValue": 0,
    }
    response = await client.post("/features/definitions", json=payload)
    assert response.status_code == 200
    data = response.json()
    assert data["name"] == "age"
    assert data["type"] == "NUMERIC"


@pytest.mark.asyncio
async def test_list_definitions(client: AsyncClient):
    response = await client.get("/features/definitions?group=user")
    assert response.status_code == 200
    data = response.json()
    assert isinstance(data, list)
    assert len(data) >= 1


@pytest.mark.asyncio
async def test_get_definition(client: AsyncClient):
    response = await client.get("/features/definitions/user/age")
    assert response.status_code == 200
    data = response.json()
    assert data["name"] == "age"
    assert data["featureGroup"] == "user"


@pytest.mark.asyncio
async def test_get_definition_not_found(client: AsyncClient):
    app.state.store.get_definition = AsyncMock(return_value=None)
    response = await client.get("/features/definitions/user/nonexistent")
    assert response.status_code == 404


@pytest.mark.asyncio
async def test_save_feature_value(client: AsyncClient):
    payload = {
        "entityId": "user-1",
        "featureName": "age",
        "featureGroup": "user",
        "value": 25,
        "version": 1,
        "computedAt": datetime.now(timezone.utc).isoformat(),
    }
    response = await client.post("/features/values", json=payload)
    assert response.status_code == 204


@pytest.mark.asyncio
async def test_get_feature_vector(client: AsyncClient):
    response = await client.get(
        "/features/vector?entityId=user-1&group=user&features=age,score"
    )
    assert response.status_code == 200
    data = response.json()
    assert data["entityId"] == "user-1"
    assert "age" in data["features"]
    assert isinstance(data["missingFeatures"], list)


@pytest.mark.asyncio
async def test_delete_entity_features(client: AsyncClient):
    response = await client.delete("/features/entity/user-1?group=user")
    assert response.status_code == 204
