"""Tests for AsyncPgStore — 10 tests using mocked asyncpg pool."""
from __future__ import annotations

import json
from datetime import datetime, timezone
from typing import Any
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from features.models import FeatureDefinition, FeatureType, FeatureValue
from features.store import AsyncPgStore, _row_to_definition, _row_to_value


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


def _mock_pool():
    pool = MagicMock()
    conn = AsyncMock()
    pool.acquire.return_value.__aenter__ = AsyncMock(return_value=conn)
    pool.acquire.return_value.__aexit__ = AsyncMock(return_value=None)
    return pool, conn


def _make_def_row(
    name: str = "feat1",
    group: str = "user",
    ftype: str = "NUMERIC",
    description: str = "",
    tags: Any = None,
    default_value: Any = None,
) -> MagicMock:
    row = MagicMock()
    row.__getitem__ = lambda self, key: {
        "name": name,
        "feature_group": group,
        "type": ftype,
        "description": description,
        "tags": json.dumps(tags or []),
        "default_value": json.dumps(default_value) if default_value is not None else None,
    }[key]
    return row


def _make_value_row(
    entity_id: str = "u1",
    feature_name: str = "feat1",
    feature_group: str = "user",
    value: Any = 42,
    version: int = 1,
) -> MagicMock:
    row = MagicMock()
    row.__getitem__ = lambda self, key: {
        "entity_id": entity_id,
        "feature_name": feature_name,
        "feature_group": feature_group,
        "value": json.dumps(value),
        "version": version,
        "computed_at": datetime.now(timezone.utc),
    }[key]
    return row


def _definition() -> FeatureDefinition:
    return FeatureDefinition(
        name="age",
        featureGroup="user",
        type=FeatureType.NUMERIC,
        description="User age",
        tags=["demographic"],
        defaultValue=0,
    )


def _feature_value() -> FeatureValue:
    return FeatureValue(
        entityId="user-123",
        featureName="age",
        featureGroup="user",
        value=30,
        version=1,
        computedAt=datetime.now(timezone.utc),
    )


# ---------------------------------------------------------------------------
# Tests
# ---------------------------------------------------------------------------


@pytest.mark.asyncio
async def test_register_feature():
    store = AsyncPgStore("postgresql://x")
    pool, conn = _mock_pool()
    store._pool = pool

    return_row = _make_def_row(name="age", group="user", ftype="NUMERIC", tags=["demographic"], default_value=0)
    conn.fetchrow = AsyncMock(return_value=return_row)

    result = await store.register_feature(_definition())
    assert result.name == "age"
    assert result.featureGroup == "user"
    assert result.type == FeatureType.NUMERIC
    conn.fetchrow.assert_called_once()


@pytest.mark.asyncio
async def test_get_definition_found():
    store = AsyncPgStore("postgresql://x")
    pool, conn = _mock_pool()
    store._pool = pool

    conn.fetchrow = AsyncMock(return_value=_make_def_row(name="age", group="user"))
    result = await store.get_definition("age", "user")
    assert result is not None
    assert result.name == "age"


@pytest.mark.asyncio
async def test_get_definition_not_found():
    store = AsyncPgStore("postgresql://x")
    pool, conn = _mock_pool()
    store._pool = pool

    conn.fetchrow = AsyncMock(return_value=None)
    result = await store.get_definition("nonexistent", "user")
    assert result is None


@pytest.mark.asyncio
async def test_save_value():
    store = AsyncPgStore("postgresql://x")
    pool, conn = _mock_pool()
    store._pool = pool

    conn.execute = AsyncMock(return_value="INSERT 0 1")
    await store.save_value(_feature_value())
    conn.execute.assert_called_once()


@pytest.mark.asyncio
async def test_get_feature_vector_all_present():
    store = AsyncPgStore("postgresql://x")
    pool, conn = _mock_pool()
    store._pool = pool

    rows = [
        _make_value_row(entity_id="u1", feature_name="age", feature_group="user", value=25),
        _make_value_row(entity_id="u1", feature_name="score", feature_group="user", value=0.9),
    ]
    conn.fetch = AsyncMock(return_value=rows)

    vector = await store.get_feature_vector("u1", ["age", "score"], "user")
    assert vector.entityId == "u1"
    assert "age" in vector.features
    assert "score" in vector.features
    assert vector.missingFeatures == []


@pytest.mark.asyncio
async def test_get_feature_vector_with_missing_uses_default():
    store = AsyncPgStore("postgresql://x")
    pool, conn = _mock_pool()
    store._pool = pool

    conn.fetch = AsyncMock(return_value=[])
    # get_definition will be called for missing feature; return one with a default
    def_row = _make_def_row(name="age", group="user", default_value=0)
    conn.fetchrow = AsyncMock(return_value=def_row)

    vector = await store.get_feature_vector("u1", ["age"], "user")
    assert "age" in vector.features
    assert vector.features["age"] == 0
    assert vector.missingFeatures == []


@pytest.mark.asyncio
async def test_get_feature_vector_truly_missing():
    store = AsyncPgStore("postgresql://x")
    pool, conn = _mock_pool()
    store._pool = pool

    conn.fetch = AsyncMock(return_value=[])
    conn.fetchrow = AsyncMock(return_value=None)

    vector = await store.get_feature_vector("u1", ["unknown_feat"], "user")
    assert "unknown_feat" not in vector.features
    assert "unknown_feat" in vector.missingFeatures


@pytest.mark.asyncio
async def test_save_batch():
    store = AsyncPgStore("postgresql://x")
    pool, conn = _mock_pool()
    store._pool = pool

    conn.execute = AsyncMock(return_value="INSERT 0 1")
    values = [
        FeatureValue(entityId="u1", featureName=f"feat{i}", featureGroup="user",
                     value=i, version=1, computedAt=datetime.now(timezone.utc))
        for i in range(3)
    ]
    await store.save_batch(values)
    assert conn.execute.call_count == 3


@pytest.mark.asyncio
async def test_get_entity_features():
    store = AsyncPgStore("postgresql://x")
    pool, conn = _mock_pool()
    store._pool = pool

    rows = [
        _make_value_row(entity_id="u1", feature_name="age", feature_group="user", value=25),
        _make_value_row(entity_id="u1", feature_name="score", feature_group="user", value=0.8),
    ]
    conn.fetch = AsyncMock(return_value=rows)

    features = await store.get_entity_features("u1", "user")
    assert len(features) == 2
    names = {f.featureName for f in features}
    assert names == {"age", "score"}


@pytest.mark.asyncio
async def test_delete_entity_features():
    store = AsyncPgStore("postgresql://x")
    pool, conn = _mock_pool()
    store._pool = pool

    conn.execute = AsyncMock(return_value="DELETE 2")
    count = await store.delete_entity_features("u1", "user")
    assert count == 2
