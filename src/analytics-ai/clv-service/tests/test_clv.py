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


def test_clv_predict_platinum(client):
    payload = {
        "customer_id": "cust-platinum",
        "orders": [
            {"order_date": "2024-01-01T00:00:00Z", "order_value": 500.0}
        ] * 20,
        "recency_days": 5,
        "frequency": 20,
        "monetary_avg": 500.0,
    }
    res = client.post("/clv/predict", json=payload)
    assert res.status_code == 200
    data = res.json()
    assert data["tier"] == "platinum"
    assert data["clv_score"] > 0
    assert data["predicted_revenue_12m"] > 0


def test_clv_predict_bronze(client):
    payload = {
        "customer_id": "cust-bronze",
        "orders": [
            {"order_date": "2023-01-01T00:00:00Z", "order_value": 20.0}
        ],
        "recency_days": 400,
        "frequency": 1,
        "monetary_avg": 20.0,
    }
    res = client.post("/clv/predict", json=payload)
    assert res.status_code == 200
    data = res.json()
    assert data["tier"] == "bronze"


def test_get_tiers(client):
    res = client.get("/clv/tiers")
    assert res.status_code == 200
    data = res.json()
    assert "tiers" in data
    assert "platinum" in data["tiers"]
    assert "gold" in data["tiers"]
    assert "silver" in data["tiers"]
    assert "bronze" in data["tiers"]
