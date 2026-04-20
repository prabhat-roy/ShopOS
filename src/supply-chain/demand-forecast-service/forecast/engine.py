from __future__ import annotations

import logging
from collections import defaultdict
from datetime import datetime, timezone
from typing import Optional

from forecast.models import ForecastResponse, InventoryAlert, SalesRecord

logger = logging.getLogger(__name__)

# Confidence thresholds (number of calendar days with sales data)
_CONFIDENCE_LOW_THRESHOLD = 7
_CONFIDENCE_MEDIUM_THRESHOLD = 30
_CONFIDENCE_LOW = 0.4
_CONFIDENCE_MEDIUM = 0.7
_CONFIDENCE_HIGH = 0.95


class ForecastEngine:
    """Computes demand forecasts using a simple moving average over daily sales."""

    def compute_forecast(
        self,
        sales_history: list[SalesRecord],
        days: int,
    ) -> ForecastResponse:
        """
        Aggregate total quantity sold per calendar day, compute the mean daily
        demand, and extrapolate over the requested horizon.

        Confidence is based on the number of distinct sale days observed:
          - < 7  days  → low   (0.40)
          - < 30 days  → medium (0.70)
          - >= 30 days → high  (0.95)
        """
        if not sales_history:
            return ForecastResponse(
                productId="unknown",
                forecastedDemand=0.0,
                averageDailySales=0.0,
                historicalDays=0,
                confidence=_CONFIDENCE_LOW,
                generatedAt=datetime.now(tz=timezone.utc),
            )

        product_id = sales_history[0].productId

        # Aggregate quantity per day
        daily_totals: dict[str, int] = defaultdict(int)
        for record in sales_history:
            day_key = record.saleDate.isoformat()
            daily_totals[day_key] += record.quantity

        distinct_days = len(daily_totals)
        total_quantity = sum(daily_totals.values())
        average_daily_sales = total_quantity / distinct_days if distinct_days > 0 else 0.0
        forecasted_demand = average_daily_sales * days

        if distinct_days < _CONFIDENCE_LOW_THRESHOLD:
            confidence = _CONFIDENCE_LOW
        elif distinct_days < _CONFIDENCE_MEDIUM_THRESHOLD:
            confidence = _CONFIDENCE_MEDIUM
        else:
            confidence = _CONFIDENCE_HIGH

        return ForecastResponse(
            productId=product_id,
            forecastedDemand=round(forecasted_demand, 4),
            averageDailySales=round(average_daily_sales, 4),
            historicalDays=distinct_days,
            confidence=confidence,
            generatedAt=datetime.now(tz=timezone.utc),
        )

    def generate_inventory_alert(
        self,
        product_id: str,
        sku: str,
        current_stock: int,
        forecast: ForecastResponse,
    ) -> Optional[InventoryAlert]:
        """
        Return an InventoryAlert when current stock is below 2x the forecasted demand.
        daysUntilStockout is the projected number of days before stock runs out
        given the average daily sales rate.
        """
        threshold = forecast.forecastedDemand * 2
        if current_stock >= threshold:
            return None

        days_until_stockout: Optional[float] = None
        if forecast.averageDailySales > 0:
            days_until_stockout = round(current_stock / forecast.averageDailySales, 2)

        return InventoryAlert(
            productId=product_id,
            sku=sku,
            currentStock=current_stock,
            forecastedDemand=forecast.forecastedDemand,
            daysUntilStockout=days_until_stockout,
        )
