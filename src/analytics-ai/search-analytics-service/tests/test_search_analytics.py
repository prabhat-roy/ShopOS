from __future__ import annotations

import pytest
from fastapi.testclient import TestClient

from main import app
import searchanalytics.api as api_module


@pytest.fixture(autouse=True)
def clear_events():
    """Reset in-memory store between tests."""
    api_module._search_events.clear()
    yield
    api_module._search_events.clear()


@pytest.fixture
def client():
    with TestClient(app) as c:
        yield c


def test_healthz(client):
    res = client.get("/healthz")
    assert res.status_code == 200
    assert res.json() == {"status": "ok"}


def test_ingest_event(client):
    res = client.post("/search-events", json={
        "query": "running shoes",
        "result_count": 42,
    })
    assert res.status_code == 201
    assert res.json()["status"] == "accepted"


def test_stats_empty(client):
    res = client.get("/search-analytics/stats")
    assert res.status_code == 200
    data = res.json()
    assert data["total_searches"] == 0
    assert data["zero_result_rate"] == 0.0


def test_stats_with_events(client):
    client.post("/search-events", json={"query": "boots", "result_count": 10})
    client.post("/search-events", json={"query": "xyz123abc", "result_count": 0})
    client.post("/search-events", json={"query": "boots", "result_count": 8})

    res = client.get("/search-analytics/stats")
    assert res.status_code == 200
    data = res.json()
    assert data["total_searches"] == 3
    assert data["zero_result_searches"] == 1
    assert data["top_queries"][0]["query"] == "boots"


def test_zero_results_endpoint(client):
    client.post("/search-events", json={"query": "noresult", "result_count": 0})
    res = client.get("/search-analytics/zero-results")
    assert res.status_code == 200
    queries = res.json()["zero_result_queries"]
    assert any(q["query"] == "noresult" for q in queries)
