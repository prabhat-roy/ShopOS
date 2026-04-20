"""
Integration tests for the recommendation-service HTTP API.

All 8 tests use the FastAPI TestClient (sync) via httpx.
The in-memory store is cleared before each test for isolation.
"""

from __future__ import annotations

import sys
import os
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

import pytest
from fastapi.testclient import TestClient

from main import app
from recommender.store import interaction_store


@pytest.fixture(autouse=True)
def clear_store():
    """Reset the in-memory store before every test."""
    interaction_store.clear()
    yield
    interaction_store.clear()


@pytest.fixture
def client() -> TestClient:
    return TestClient(app)


# ---------------------------------------------------------------------------
# Test 1: GET /healthz returns 200 and status=ok
# ---------------------------------------------------------------------------


def test_healthz(client: TestClient):
    response = client.get("/healthz")
    assert response.status_code == 200
    body = response.json()
    assert body["status"] == "ok"
    assert "total_interactions" in body


# ---------------------------------------------------------------------------
# Test 2: POST /interactions records an interaction successfully
# ---------------------------------------------------------------------------


def test_record_interaction(client: TestClient):
    payload = {
        "userId": "user-1",
        "productId": "prod-A",
        "interactionType": "purchase",
        "score": 1.0,
    }
    response = client.post("/interactions", json=payload)
    assert response.status_code == 201
    body = response.json()
    assert body["recorded"] is True
    assert body["userId"] == "user-1"
    assert body["productId"] == "prod-A"


# ---------------------------------------------------------------------------
# Test 3: GET /interactions/user/{userId} returns recorded interactions
# ---------------------------------------------------------------------------


def test_get_user_interactions(client: TestClient):
    # Seed an interaction
    client.post(
        "/interactions",
        json={"userId": "user-2", "productId": "prod-B", "interactionType": "view", "score": 1.0},
    )
    response = client.get("/interactions/user/user-2")
    assert response.status_code == 200
    data = response.json()
    assert isinstance(data, list)
    assert len(data) == 1
    assert data[0]["productId"] == "prod-B"


# ---------------------------------------------------------------------------
# Test 4: GET /interactions/user/{userId} for unknown user returns 404
# ---------------------------------------------------------------------------


def test_get_user_interactions_not_found(client: TestClient):
    response = client.get("/interactions/user/no-such-user")
    assert response.status_code == 404


# ---------------------------------------------------------------------------
# Test 5: GET /recommendations/popular returns list (even with no data)
# ---------------------------------------------------------------------------


def test_popular_empty_store(client: TestClient):
    response = client.get("/recommendations/popular")
    assert response.status_code == 200
    assert response.json() == []


# ---------------------------------------------------------------------------
# Test 6: GET /recommendations/popular after seeding returns products
# ---------------------------------------------------------------------------


def test_popular_with_data(client: TestClient):
    for pid in ("prod-1", "prod-2", "prod-3"):
        client.post(
            "/interactions",
            json={"userId": "u1", "productId": pid, "interactionType": "purchase", "score": 1.0},
        )

    response = client.get("/recommendations/popular?limit=2")
    assert response.status_code == 200
    body = response.json()
    assert len(body) <= 2
    for item in body:
        assert "productId" in item
        assert "score" in item


# ---------------------------------------------------------------------------
# Test 7: POST /recommendations with strategy=popular
# ---------------------------------------------------------------------------


def test_post_recommendations_popular(client: TestClient):
    client.post(
        "/interactions",
        json={"userId": "u-x", "productId": "prod-X", "interactionType": "view", "score": 1.0},
    )
    payload = {"strategy": "popular", "limit": 5}
    response = client.post("/recommendations", json=payload)
    assert response.status_code == 200
    body = response.json()
    assert "recommendations" in body
    assert body["strategy"] == "popular"


# ---------------------------------------------------------------------------
# Test 8: POST /recommendations with userId triggers user-based or hybrid
# ---------------------------------------------------------------------------


def test_post_recommendations_with_user(client: TestClient):
    # Seed two users with overlapping products
    for pid in ("p1", "p2", "p3"):
        client.post(
            "/interactions",
            json={"userId": "alice", "productId": pid, "interactionType": "purchase", "score": 1.0},
        )
    for pid in ("p1", "p2", "p4"):
        client.post(
            "/interactions",
            json={"userId": "bob", "productId": pid, "interactionType": "purchase", "score": 1.0},
        )

    payload = {"userId": "alice", "strategy": "hybrid", "limit": 5}
    response = client.post("/recommendations", json=payload)
    assert response.status_code == 200
    body = response.json()
    assert body["userId"] == "alice"
    assert isinstance(body["recommendations"], list)
    # alice already bought p1,p2,p3 — they must not appear
    rec_ids = {r["productId"] for r in body["recommendations"]}
    for seen_pid in ("p1", "p2", "p3"):
        assert seen_pid not in rec_ids, f"{seen_pid} should not be recommended to alice"
