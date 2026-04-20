from __future__ import annotations

from datetime import datetime
from typing import Optional
from pydantic import BaseModel, Field


class PageViewEvent(BaseModel):
    sessionId: str
    userId: Optional[str] = None
    pageUrl: str
    referrer: Optional[str] = None
    userAgent: Optional[str] = None
    timestamp: datetime = Field(default_factory=datetime.utcnow)


class ProductClickEvent(BaseModel):
    sessionId: str
    userId: Optional[str] = None
    productId: str
    sku: str
    position: Optional[int] = None
    timestamp: datetime = Field(default_factory=datetime.utcnow)


class SearchEvent(BaseModel):
    sessionId: str
    userId: Optional[str] = None
    query: str
    resultCount: int = 0
    timestamp: datetime = Field(default_factory=datetime.utcnow)


class AnalyticsQuery(BaseModel):
    startDate: datetime
    endDate: datetime
    metric: str
    groupBy: Optional[str] = None
    limit: int = 100


class PageViewRecord(BaseModel):
    sessionId: str
    userId: Optional[str] = None
    pageUrl: str
    referrer: Optional[str] = None
    userAgent: Optional[str] = None
    timestamp: datetime


class ProductClickRecord(BaseModel):
    sessionId: str
    userId: Optional[str] = None
    productId: str
    sku: str
    position: Optional[int] = None
    clickCount: int = 1
    timestamp: datetime


class SearchRecord(BaseModel):
    sessionId: str
    userId: Optional[str] = None
    query: str
    resultCount: int
    timestamp: datetime


class TopProduct(BaseModel):
    productId: str
    sku: str
    clickCount: int


class SearchQueryStat(BaseModel):
    query: str
    searchCount: int
    avgResultCount: float


class AnalyticsSummary(BaseModel):
    totalPageViews: int
    totalProductClicks: int
    totalSearches: int
    uniqueSessions: int
    startDate: datetime
    endDate: datetime
