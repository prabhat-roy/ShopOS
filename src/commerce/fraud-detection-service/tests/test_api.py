"""Integration-style tests for the fraud-detection API (no real DB)."""

from __future__ import annotations

import json
from datetime import datetime, timezone
from unittest.mock import AsyncMock, MagicMock

import pytest
from fastapi.testclient import TestClient

from fraud.models import FraudCheckResult
from main import app


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _make_result(order_id: str = "order-001") -> FraudCheckResult:
    return FraudCheckResult(
        order_id=order_id,
        risk_score=10,
        risk_level="low",
        decision="approve",
        signals=["No fraud signals detected"],
        checked_at=datetime.now(timezone.utc),
    )


def _mock_store(result: FraudCheckResult | None = None) -> MagicMock:
    store = MagicMock()
    store.save_result = AsyncMock(return_value=None)
    store.get_result = AsyncMock(return_value=result)
    store.list_flagged = AsyncMock(return_value=[result] if result else [])
    return store


# ---------------------------------------------------------------------------
# Tests
# ---------------------------------------------------------------------------

class TestFraudAPI:
    def setup_method(self):
        self.client = TestClient(app, raise_server_exceptions=True)

    def _patch_store(self, store):
        app.state.store = store

    def test_healthz(self):
        response = self.client.get("/healthz")
        assert response.status_code == 200
        assert response.json() == {"status": "ok"}

    def test_check_fraud_returns_fraud_check_result_shape(self):
        mock_store = _mock_store(_make_result())
        self._patch_store(mock_store)

        payload = {
            "order_id": "order-001",
            "customer_id": "cust-001",
            "amount": 100.0,
            "currency": "USD",
            "ip_address": "1.2.3.4",
            "device_fingerprint": "fp-xyz",
            "items": [],
            "shipping_address": {"country": "US", "postal_code": "10001"},
            "billing_address": {"country": "US", "postal_code": "10001"},
        }
        response = self.client.post("/fraud/check", json=payload)
        assert response.status_code == 200

        body = response.json()
        assert body["order_id"] == "order-001"
        assert "risk_score" in body
        assert body["risk_level"] in ("low", "medium", "high", "critical")
        assert body["decision"] in ("approve", "review", "decline")
        assert isinstance(body["signals"], list)
        assert "checked_at" in body

    def test_get_result_returns_404_when_not_found(self):
        mock_store = _mock_store(None)
        self._patch_store(mock_store)

        response = self.client.get("/fraud/results/nonexistent-order")
        assert response.status_code == 404

    def test_get_result_returns_existing_result(self):
        result = _make_result("order-999")
        mock_store = _mock_store(result)
        self._patch_store(mock_store)

        response = self.client.get("/fraud/results/order-999")
        assert response.status_code == 200
        body = response.json()
        assert body["order_id"] == "order-999"
        assert body["risk_level"] == "low"

    def test_list_flagged_returns_list(self):
        result = FraudCheckResult(
            order_id="order-flagged",
            risk_score=70,
            risk_level="high",
            decision="review",
            signals=["High-value order"],
            checked_at=datetime.now(timezone.utc),
        )
        mock_store = _mock_store(result)
        self._patch_store(mock_store)

        response = self.client.get("/fraud/flagged?limit=10")
        assert response.status_code == 200
        body = response.json()
        assert isinstance(body, list)
        assert body[0]["decision"] == "review"
