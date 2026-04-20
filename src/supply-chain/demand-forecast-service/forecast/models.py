from __future__ import annotations

from datetime import date, datetime
from typing import Optional

from pydantic import BaseModel, Field


class SalesRecord(BaseModel):
    productId: str
    sku: str
    quantity: int
    saleDate: date
    orderId: str


class ForecastRequest(BaseModel):
    productId: str
    days: int = Field(default=30, ge=1, le=365)


class ForecastResponse(BaseModel):
    productId: str
    forecastedDemand: float
    averageDailySales: float
    historicalDays: int
    confidence: float
    generatedAt: datetime


class InventoryAlert(BaseModel):
    productId: str
    sku: str
    currentStock: int
    forecastedDemand: float
    daysUntilStockout: Optional[float] = None
