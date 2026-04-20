"""
Integration tests for the sentiment-analysis-service HTTP API.

The PostgreSQL store is mocked so tests run without a live database.
All 8 tests use FastAPI TestClient.
"""

from __future__ import annotations

import sys
import os
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from datetime import datetime, timezone
from unittest.mock import AsyncMock, MagicMock, patch

import pytest
from fastapi.testclient import TestClient

from sentiment.models import AggregateStats, SentimentLabel, SentimentResult

# Build the result fixture before importing app so the mock is ready
_MOCK_RESULT = SentimentResult(
    text="This is an amazing product!",
    label=SentimentLabel.POSITIVE,
    score=0.857,
    positiveWords=["amazing"],
    negativeWords=[],
    entityId="review-001",
    entityType="review",
    analyzedAt=datetime(2024, 1, 1, tzinfo=timezone.utc),
)


def _make_mock_store() -> MagicMock:
    store = MagicMock()
    store.save_result = AsyncMock(return_value=1)
    store.get_result = AsyncMock(return_value=_MOCK_RESULT)
    store.list_results = AsyncMock(return_value=[_MOCK_RESULT])
    store.get_aggregate_stats = AsyncMock(
        return_value=AggregateStats(
            entityType="review",
            positive=10,
            negative=3,
            neutral=5,
            total=18,
            avgScore=0.71,
        )
    )
    store.close = AsyncMock()
    return store


@pytest.fixture
def client() -> TestClient:
    from main import app

    mock_store = _make_mock_store()
    app.state.store = mock_store
    return TestClient(app)


# ---------------------------------------------------------------------------
# Test 1: GET /healthz
# ---------------------------------------------------------------------------


def test_healthz(client: TestClient):
    response = client.get("/healthz")
    assert response.status_code == 200
    assert response.json()["status"] == "ok"


# ---------------------------------------------------------------------------
# Test 2: POST /sentiment/analyze returns SentimentResult
# ---------------------------------------------------------------------------


def test_analyze_single(client: TestClient):
    payload = {
        "text": "This is an amazing product!",
        "entityId": "review-001",
        "entityType": "review",
    }
    response = client.post("/sentiment/analyze", json=payload)
    assert response.status_code == 200
    body = response.json()
    assert "label" in body
    assert "score" in body
    assert body["label"] == SentimentLabel.POSITIVE.value
    assert 0.0 <= body["score"] <= 1.0


# ---------------------------------------------------------------------------
# Test 3: POST /sentiment/analyze negative text
# ---------------------------------------------------------------------------


def test_analyze_negative(client: TestClient):
    payload = {"text": "Terrible quality, broken on arrival, awful experience!"}
    response = client.post("/sentiment/analyze", json=payload)
    assert response.status_code == 200
    body = response.json()
    assert body["label"] == SentimentLabel.NEGATIVE.value


# ---------------------------------------------------------------------------
# Test 4: POST /sentiment/batch returns list of results
# ---------------------------------------------------------------------------


def test_batch_analyze(client: TestClient):
    payload = {
        "texts": [
            "Amazing product, highly recommend!",
            "Terrible, broken, complete waste.",
            "Arrived on Wednesday.",
        ]
    }
    response = client.post("/sentiment/batch", json=payload)
    assert response.status_code == 200
    body = response.json()
    assert isinstance(body, list)
    assert len(body) == 3
    for item in body:
        assert "label" in item
        assert "score" in item


# ---------------------------------------------------------------------------
# Test 5: POST /sentiment/batch rejects more than 50 texts
# ---------------------------------------------------------------------------


def test_batch_too_many_texts(client: TestClient):
    payload = {"texts": ["some text"] * 51}
    response = client.post("/sentiment/batch", json=payload)
    assert response.status_code == 422


# ---------------------------------------------------------------------------
# Test 6: GET /sentiment/{entityId} returns stored result
# ---------------------------------------------------------------------------


def test_get_by_entity(client: TestClient):
    response = client.get("/sentiment/review-001")
    assert response.status_code == 200
    body = response.json()
    assert body["entityId"] == "review-001"


# ---------------------------------------------------------------------------
# Test 7: GET /sentiment?entityType=review returns list
# ---------------------------------------------------------------------------


def test_list_results(client: TestClient):
    response = client.get("/sentiment?entityType=review&limit=10")
    assert response.status_code == 200
    body = response.json()
    assert isinstance(body, list)
    assert len(body) >= 1


# ---------------------------------------------------------------------------
# Test 8: GET /sentiment/stats returns aggregate stats
# ---------------------------------------------------------------------------


def test_get_stats(client: TestClient):
    response = client.get("/sentiment/stats?entityType=review")
    assert response.status_code == 200
    body = response.json()
    assert "positive" in body
    assert "negative" in body
    assert "neutral" in body
    assert "total" in body
    assert "avgScore" in body
    assert body["positive"] + body["negative"] + body["neutral"] == body["total"]
