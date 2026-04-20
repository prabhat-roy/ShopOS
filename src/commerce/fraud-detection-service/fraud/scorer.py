from datetime import datetime, timezone

from fraud.config import settings
from fraud.models import FraudCheckRequest, FraudCheckResult

# High-risk country ISO codes (illustrative list)
_HIGH_RISK_COUNTRIES = {"NG", "GH", "KP", "IR", "IQ", "LY", "SO", "SD", "SY", "YE"}

# In-memory IP order counter (resets on restart; real impl would use Redis)
_ip_order_counts: dict[str, int] = {}


class FraudScorer:
    """Rule-based fraud scorer.

    Score bands:
        0-30   → low     / approve
        31-60  → medium  / review
        61-80  → high    / review
        81-100 → critical / decline
    """

    def score(self, req: FraudCheckRequest) -> FraudCheckResult:
        score = 0
        signals: list[str] = []

        # Rule 1: High-value order (+30)
        if req.amount > 5000:
            score += 30
            signals.append(f"Order amount ${req.amount:.2f} exceeds $5,000 threshold")

        # Rule 2: New / unknown device fingerprint (+10)
        if not req.device_fingerprint:
            score += 10
            signals.append("No device fingerprint provided")

        # Rule 3: Multiple orders from the same IP in this session (+20)
        if req.ip_address:
            count = _ip_order_counts.get(req.ip_address, 0) + 1
            _ip_order_counts[req.ip_address] = count
            if count > 1:
                score += 20
                signals.append(
                    f"IP address {req.ip_address} has {count} orders in current window"
                )

        # Rule 4: High-risk shipping country (+25)
        country = (req.shipping_address or {}).get("country", "").upper()
        if country in _HIGH_RISK_COUNTRIES:
            score += 25
            signals.append(f"Shipping destination country '{country}' is high-risk")

        # Rule 5: Billing / shipping address mismatch (+15)
        if req.shipping_address and req.billing_address:
            ship_country = (req.shipping_address.get("country") or "").upper()
            bill_country = (req.billing_address.get("country") or "").upper()
            ship_postal = req.shipping_address.get("postal_code", "")
            bill_postal = req.billing_address.get("postal_code", "")
            if ship_country != bill_country or ship_postal != bill_postal:
                score += 15
                signals.append("Shipping address does not match billing address")

        # Clamp score to 0-100
        score = min(score, 100)

        # Determine risk level and decision
        threshold_high = settings.risk_threshold_high        # default 60
        threshold_critical = settings.risk_threshold_critical  # default 80

        if score <= 30:
            risk_level = "low"
            decision = "approve"
        elif score <= threshold_high:
            risk_level = "medium"
            decision = "review"
        elif score <= threshold_critical:
            risk_level = "high"
            decision = "review"
        else:
            risk_level = "critical"
            decision = "decline"

        if not signals:
            signals.append("No fraud signals detected")

        return FraudCheckResult(
            order_id=req.order_id,
            risk_score=score,
            risk_level=risk_level,
            decision=decision,
            signals=signals,
            checked_at=datetime.now(timezone.utc),
        )
