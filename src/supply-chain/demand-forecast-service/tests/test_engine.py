"""Tests for ForecastEngine — pure unit tests, no I/O required."""

from __future__ import annotations

from datetime import date, timedelta

import pytest

from forecast.engine import (
    ForecastEngine,
    _CONFIDENCE_HIGH,
    _CONFIDENCE_LOW,
    _CONFIDENCE_MEDIUM,
)
from forecast.models import ForecastResponse, SalesRecord


def _make_record(product_id: str, sku: str, quantity: int, days_ago: int, order_id: str) -> SalesRecord:
    return SalesRecord(
        productId=product_id,
        sku=sku,
        quantity=quantity,
        saleDate=date.today() - timedelta(days=days_ago),
        orderId=order_id,
    )


engine = ForecastEngine()


# ------------------------------------------------------------------
# 1. Empty history returns zero demand
# ------------------------------------------------------------------
def test_empty_history_returns_zero_demand() -> None:
    result = engine.compute_forecast([], days=30)
    assert result.forecastedDemand == 0.0
    assert result.averageDailySales == 0.0
    assert result.historicalDays == 0


# ------------------------------------------------------------------
# 2. Seven-day history computes correct moving average
# ------------------------------------------------------------------
def test_seven_day_history_computes_average() -> None:
    records = [
        _make_record("prod-1", "SKU-1", 10, days_ago=i, order_id=f"ord-{i}")
        for i in range(7)
    ]
    result = engine.compute_forecast(records, days=7)
    # 7 days * 10 units/day = 70
    assert result.forecastedDemand == pytest.approx(70.0, rel=1e-3)
    assert result.averageDailySales == pytest.approx(10.0, rel=1e-3)
    assert result.historicalDays == 7


# ------------------------------------------------------------------
# 3. Confidence is LOW when fewer than 7 distinct sale days
# ------------------------------------------------------------------
def test_confidence_low_when_fewer_than_7_days() -> None:
    records = [
        _make_record("prod-2", "SKU-2", 5, days_ago=i, order_id=f"ord-low-{i}")
        for i in range(3)
    ]
    result = engine.compute_forecast(records, days=30)
    assert result.confidence == _CONFIDENCE_LOW


# ------------------------------------------------------------------
# 4. Confidence is MEDIUM between 7 and 29 distinct sale days
# ------------------------------------------------------------------
def test_confidence_medium_between_7_and_29_days() -> None:
    records = [
        _make_record("prod-3", "SKU-3", 2, days_ago=i, order_id=f"ord-med-{i}")
        for i in range(15)
    ]
    result = engine.compute_forecast(records, days=30)
    assert result.confidence == _CONFIDENCE_MEDIUM


# ------------------------------------------------------------------
# 5. Confidence is HIGH when 30 or more distinct sale days
# ------------------------------------------------------------------
def test_confidence_high_when_30_or_more_days() -> None:
    records = [
        _make_record("prod-4", "SKU-4", 8, days_ago=i, order_id=f"ord-high-{i}")
        for i in range(30)
    ]
    result = engine.compute_forecast(records, days=30)
    assert result.confidence == _CONFIDENCE_HIGH


# ------------------------------------------------------------------
# 6. Multiple orders on same day are aggregated correctly
# ------------------------------------------------------------------
def test_same_day_quantities_are_aggregated() -> None:
    today = date.today()
    records = [
        SalesRecord(productId="prod-5", sku="SKU-5", quantity=4, saleDate=today, orderId="ord-a"),
        SalesRecord(productId="prod-5", sku="SKU-5", quantity=6, saleDate=today, orderId="ord-b"),
    ]
    result = engine.compute_forecast(records, days=1)
    # 1 unique day, total qty = 10 → avg daily = 10, forecast over 1 day = 10
    assert result.averageDailySales == pytest.approx(10.0)
    assert result.forecastedDemand == pytest.approx(10.0)
    assert result.historicalDays == 1


# ------------------------------------------------------------------
# 7. Inventory alert triggered when stock < 2x forecast
# ------------------------------------------------------------------
def test_inventory_alert_triggered_below_threshold() -> None:
    forecast = ForecastResponse(
        productId="prod-6",
        forecastedDemand=100.0,
        averageDailySales=10.0,
        historicalDays=10,
        confidence=_CONFIDENCE_MEDIUM,
        generatedAt=__import__("datetime").datetime.utcnow(),
    )
    alert = engine.generate_inventory_alert("prod-6", "SKU-6", current_stock=150, forecast=forecast)
    assert alert is not None
    assert alert.productId == "prod-6"
    assert alert.currentStock == 150
    # daysUntilStockout = 150 / 10 = 15
    assert alert.daysUntilStockout == pytest.approx(15.0)


# ------------------------------------------------------------------
# 8. Inventory alert NOT triggered when stock >= 2x forecast
# ------------------------------------------------------------------
def test_inventory_alert_not_triggered_above_threshold() -> None:
    forecast = ForecastResponse(
        productId="prod-7",
        forecastedDemand=100.0,
        averageDailySales=10.0,
        historicalDays=10,
        confidence=_CONFIDENCE_MEDIUM,
        generatedAt=__import__("datetime").datetime.utcnow(),
    )
    alert = engine.generate_inventory_alert("prod-7", "SKU-7", current_stock=200, forecast=forecast)
    assert alert is None
