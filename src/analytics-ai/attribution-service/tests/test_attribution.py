from __future__ import annotations

import pytest
from fastapi.testclient import TestClient

from main import app


@pytest.fixture
def client():
    with TestClient(app) as c:
        yield c


def test_healthz(client):
    res = client.get("/healthz")
    assert res.status_code == 200
    assert res.json() == {"status": "ok"}


def test_attribution_linear(client):
    touchpoints = [
        {"channel": "email", "timestamp": "2024-01-01T10:00:00Z"},
        {"channel": "social", "timestamp": "2024-01-02T10:00:00Z"},
        {"channel": "direct", "timestamp": "2024-01-03T10:00:00Z"},
    ]
    res = client.post(
        "/attribution/calculate",
        params={"customer_id": "cust-1", "model": "linear", "conversion_value": 90.0},
        json=touchpoints,
    )
    assert res.status_code == 200
    data = res.json()
    assert data["model"] == "linear"
    assert "email" in data["attributions"]


def test_attribution_first_click(client):
    touchpoints = [
        {"channel": "email", "timestamp": "2024-01-01T10:00:00Z"},
        {"channel": "social", "timestamp": "2024-01-02T10:00:00Z"},
    ]
    res = client.post(
        "/attribution/calculate",
        params={"customer_id": "cust-2", "model": "first_click", "conversion_value": 100.0},
        json=touchpoints,
    )
    assert res.status_code == 200
    data = res.json()
    assert data["attributions"]["email"] == 100.0


def test_attribution_last_click(client):
    touchpoints = [
        {"channel": "email", "timestamp": "2024-01-01T10:00:00Z"},
        {"channel": "paid_search", "timestamp": "2024-01-02T10:00:00Z"},
    ]
    res = client.post(
        "/attribution/calculate",
        params={"customer_id": "cust-3", "model": "last_click", "conversion_value": 50.0},
        json=touchpoints,
    )
    assert res.status_code == 200
    assert res.json()["attributions"]["paid_search"] == 50.0
