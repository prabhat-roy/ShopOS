"""Unit tests for the fraud scorer."""

import pytest

from fraud.models import FraudCheckRequest
from fraud.scorer import FraudScorer


@pytest.fixture()
def scorer() -> FraudScorer:
    return FraudScorer()


def _base_req(**kwargs) -> FraudCheckRequest:
    defaults = dict(
        order_id="order-001",
        customer_id="cust-001",
        amount=100.0,
        currency="USD",
        ip_address="",
        device_fingerprint="fp-abc",
        items=[],
        shipping_address={"country": "US", "postal_code": "10001"},
        billing_address={"country": "US", "postal_code": "10001"},
    )
    defaults.update(kwargs)
    return FraudCheckRequest(**defaults)


class TestFraudScorer:
    def test_low_risk_clean_order(self, scorer: FraudScorer):
        req = _base_req(amount=50.0, device_fingerprint="fp-abc")
        result = scorer.score(req)
        assert result.risk_level == "low"
        assert result.decision == "approve"
        assert result.risk_score <= 30

    def test_high_amount_raises_score(self, scorer: FraudScorer):
        req = _base_req(amount=6000.0, device_fingerprint="fp-abc")
        result = scorer.score(req)
        # +30 for amount > 5000 → at least medium
        assert result.risk_score >= 30
        assert result.risk_level in ("medium", "high", "critical")

    def test_missing_device_fingerprint_adds_signal(self, scorer: FraudScorer):
        req = _base_req(amount=50.0, device_fingerprint="")
        result = scorer.score(req)
        assert any("fingerprint" in s.lower() for s in result.signals)

    def test_high_risk_country_raises_score(self, scorer: FraudScorer):
        req = _base_req(
            amount=50.0,
            device_fingerprint="fp-abc",
            shipping_address={"country": "NG", "postal_code": "100001"},
            billing_address={"country": "NG", "postal_code": "100001"},
        )
        result = scorer.score(req)
        assert result.risk_score >= 25
        assert any("high-risk" in s.lower() for s in result.signals)

    def test_address_mismatch_adds_signal(self, scorer: FraudScorer):
        req = _base_req(
            amount=50.0,
            device_fingerprint="fp-abc",
            shipping_address={"country": "CA", "postal_code": "M5V"},
            billing_address={"country": "US", "postal_code": "10001"},
        )
        result = scorer.score(req)
        assert any("billing" in s.lower() or "shipping" in s.lower() for s in result.signals)

    def test_critical_high_value_in_high_risk_country(self, scorer: FraudScorer):
        req = _base_req(
            order_id="order-critical",
            amount=6000.0,
            device_fingerprint="",
            ip_address="",
            shipping_address={"country": "NG", "postal_code": "100001"},
            billing_address={"country": "US", "postal_code": "10001"},
        )
        result = scorer.score(req)
        # amount +30, no fingerprint +10, high-risk country +25, mismatch +15 = 80 → critical
        assert result.risk_score >= 80
        assert result.risk_level == "critical"
        assert result.decision == "decline"

    def test_result_has_required_fields(self, scorer: FraudScorer):
        req = _base_req()
        result = scorer.score(req)
        assert result.order_id == "order-001"
        assert isinstance(result.risk_score, int)
        assert result.risk_level in ("low", "medium", "high", "critical")
        assert result.decision in ("approve", "review", "decline")
        assert isinstance(result.signals, list)
        assert result.checked_at is not None
