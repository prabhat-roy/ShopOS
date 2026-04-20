from __future__ import annotations

from enum import Enum
from typing import List, Optional

from fastapi import APIRouter
from pydantic import BaseModel

router = APIRouter()


class CustomerTier(str, Enum):
    PLATINUM = "platinum"
    GOLD = "gold"
    SILVER = "silver"
    BRONZE = "bronze"


class OrderFeature(BaseModel):
    order_date: str
    order_value: float
    category: Optional[str] = None


class CLVRequest(BaseModel):
    customer_id: str
    orders: List[OrderFeature]
    recency_days: int = 0
    frequency: int = 0
    monetary_avg: float = 0.0


class CLVResponse(BaseModel):
    customer_id: str
    clv_score: float
    tier: CustomerTier
    predicted_revenue_12m: float
    rfm_recency: int
    rfm_frequency: int
    rfm_monetary: float


@router.get("/healthz")
async def health():
    return {"status": "ok"}


@router.post("/clv/predict", response_model=CLVResponse)
async def predict_clv(payload: CLVRequest):
    """
    Predict Customer Lifetime Value and assign a tier segment.
    Uses RFM (Recency, Frequency, Monetary) features to compute a CLV score.
    """
    clv_score, predicted_12m = _compute_clv(
        recency_days=payload.recency_days,
        frequency=payload.frequency,
        monetary_avg=payload.monetary_avg,
        orders=payload.orders,
    )
    tier = _assign_tier(clv_score)

    return CLVResponse(
        customer_id=payload.customer_id,
        clv_score=round(clv_score, 2),
        tier=tier,
        predicted_revenue_12m=round(predicted_12m, 2),
        rfm_recency=payload.recency_days,
        rfm_frequency=payload.frequency,
        rfm_monetary=payload.monetary_avg,
    )


@router.get("/clv/tiers")
async def get_tiers():
    """Returns tier thresholds used for customer segmentation."""
    return {
        "tiers": {
            "platinum": {"min_clv": 5000, "description": "Top 5% — highest value customers"},
            "gold": {"min_clv": 2000, "description": "Top 10–20% — high value customers"},
            "silver": {"min_clv": 500, "description": "Mid-tier customers"},
            "bronze": {"min_clv": 0, "description": "Entry-level customers"},
        }
    }


# ---------------------------------------------------------------------------
# CLV computation (simplified RFM-based model)
# ---------------------------------------------------------------------------

def _compute_clv(
    recency_days: int,
    frequency: int,
    monetary_avg: float,
    orders: list,
) -> tuple[float, float]:
    """
    Simplified CLV model using RFM features.
    Real implementation would use BG/NBD + Gamma-Gamma or an ML model.
    """
    # Recency score: lower days = higher score (max 100 at 0 days, min 0 at 365+)
    recency_score = max(0.0, 100.0 - (recency_days / 365.0) * 100.0)

    # Frequency score: capped at 50 orders
    frequency_score = min(frequency * 2.0, 100.0)

    # Monetary score: based on average order value
    monetary_score = min(monetary_avg / 10.0, 100.0)

    # Weighted CLV score
    clv_score = (recency_score * 0.3) + (frequency_score * 0.4) + (monetary_score * 0.3)

    # Predicted 12-month revenue: purchase rate × avg order value × 12 months
    purchase_rate = frequency / 12.0 if frequency > 0 else 0.5
    predicted_12m = purchase_rate * monetary_avg * 12.0

    return clv_score, predicted_12m


def _assign_tier(clv_score: float) -> CustomerTier:
    if clv_score >= 80:
        return CustomerTier.PLATINUM
    elif clv_score >= 60:
        return CustomerTier.GOLD
    elif clv_score >= 35:
        return CustomerTier.SILVER
    return CustomerTier.BRONZE
