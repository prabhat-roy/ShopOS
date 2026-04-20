from datetime import datetime
from typing import Optional
from pydantic import BaseModel, Field


class PricePoint(BaseModel):
    price: float
    demand: float


class PriceOptimizationRequest(BaseModel):
    productId: str
    currentPrice: float
    costPrice: float
    minPrice: float
    maxPrice: float
    competitorPrices: list[float] = Field(default_factory=list)
    targetMargin: float = 0.3
    elasticity: float = -1.5


class PriceOptimizationResult(BaseModel):
    productId: str
    currentPrice: float
    suggestedPrice: float
    expectedDemandChange: float
    expectedRevenueChange: float
    marginAtSuggestedPrice: float
    confidence: float
    reasoning: str
    generatedAt: datetime


class PricingRule(BaseModel):
    productId: str
    minPrice: float
    maxPrice: float
    targetMargin: float
    active: bool = True
