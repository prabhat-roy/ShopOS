from __future__ import annotations

from enum import Enum
from typing import List

from fastapi import APIRouter
from pydantic import BaseModel

router = APIRouter()


class AttributionModel(str, Enum):
    FIRST_CLICK = "first_click"
    LAST_CLICK = "last_click"
    LINEAR = "linear"
    TIME_DECAY = "time_decay"


class Touchpoint(BaseModel):
    channel: str
    timestamp: str
    campaign: str | None = None


class AttributionResult(BaseModel):
    customer_id: str
    conversion_value: float
    model: AttributionModel
    attributions: dict


@router.get("/healthz")
async def health():
    return {"status": "ok"}


@router.post("/attribution/calculate", response_model=AttributionResult)
async def calculate_attribution(
    customer_id: str,
    touchpoints: List[Touchpoint],
    model: AttributionModel = AttributionModel.LINEAR,
    conversion_value: float = 0.0,
):
    """
    Calculate marketing attribution for a customer conversion journey.
    Supports first-click, last-click, linear, and time-decay models.
    """
    attributions = _calculate(touchpoints, model, conversion_value)
    return AttributionResult(
        customer_id=customer_id,
        conversion_value=conversion_value,
        model=model,
        attributions=attributions,
    )


def _calculate(
    touchpoints: List[Touchpoint],
    model: AttributionModel,
    value: float,
) -> dict:
    if not touchpoints:
        return {}

    channels = [tp.channel for tp in touchpoints]
    n = len(channels)
    result: dict = {}

    if model == AttributionModel.FIRST_CLICK:
        result[channels[0]] = value

    elif model == AttributionModel.LAST_CLICK:
        result[channels[-1]] = value

    elif model == AttributionModel.LINEAR:
        share = round(value / n, 4) if n else 0
        for ch in channels:
            result[ch] = result.get(ch, 0) + share

    elif model == AttributionModel.TIME_DECAY:
        # Exponential weights: most recent touchpoint gets highest weight
        import math
        weights = [math.exp(0.5 * i) for i in range(n)]
        total = sum(weights)
        for ch, w in zip(channels, weights):
            result[ch] = result.get(ch, 0) + round(value * w / total, 4)

    return result
