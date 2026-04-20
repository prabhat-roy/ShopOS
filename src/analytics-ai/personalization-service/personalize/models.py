from datetime import datetime
from typing import Optional

from pydantic import BaseModel, Field


class UserProfile(BaseModel):
    userId: str
    preferredCategories: list[str] = Field(default_factory=list)
    preferredBrands: list[str] = Field(default_factory=list)
    priceRangeLow: float = 0.0
    priceRangeHigh: float = 10000.0
    recentlyViewedProducts: list[str] = Field(default_factory=list, max_length=50)
    purchaseHistory: list[str] = Field(default_factory=list)
    excludedCategories: list[str] = Field(default_factory=list)
    updatedAt: datetime = Field(default_factory=datetime.utcnow)


class PersonalizationRequest(BaseModel):
    userId: str
    contextType: str  # "homepage" | "category" | "search" | "cart"
    candidateProductIds: list[str]
    limit: int = 20


class PersonalizedResult(BaseModel):
    userId: str
    rankedProductIds: list[str]
    scores: dict[str, float]
    contextType: str
    profile: Optional[UserProfile] = None
