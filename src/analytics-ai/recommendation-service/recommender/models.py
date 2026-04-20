from __future__ import annotations

from datetime import datetime, timezone
from typing import Literal, Optional

from pydantic import BaseModel, Field, field_validator


class UserInteraction(BaseModel):
    userId: str = Field(..., min_length=1)
    productId: str = Field(..., min_length=1)
    interactionType: Literal["view", "purchase", "wishlist", "cart"]
    score: float = Field(default=1.0, ge=0.0, le=10.0)
    timestamp: datetime = Field(default_factory=lambda: datetime.now(timezone.utc))


class ProductRec(BaseModel):
    productId: str
    score: float = Field(..., ge=0.0)
    reason: str


class RecommendRequest(BaseModel):
    userId: Optional[str] = None
    productId: Optional[str] = None
    strategy: str = Field(default="hybrid")
    limit: int = Field(default=10, ge=1, le=100)

    @field_validator("strategy")
    @classmethod
    def validate_strategy(cls, v: str) -> str:
        allowed = {"hybrid", "user-based", "item-based", "popular"}
        if v not in allowed:
            raise ValueError(f"strategy must be one of {allowed}")
        return v


class RecommendResponse(BaseModel):
    recommendations: list[ProductRec]
    strategy: str
    userId: Optional[str] = None
    productId: Optional[str] = None


class InteractionRecordedResponse(BaseModel):
    recorded: bool = True
    userId: str
    productId: str
    interactionType: str
