from __future__ import annotations

from datetime import datetime, timezone
from enum import Enum
from typing import Optional

from pydantic import BaseModel, Field, field_validator


class SentimentLabel(str, Enum):
    POSITIVE = "POSITIVE"
    NEGATIVE = "NEGATIVE"
    NEUTRAL = "NEUTRAL"


class AnalyzeRequest(BaseModel):
    text: str = Field(..., min_length=1, max_length=10000)
    entityId: Optional[str] = None
    entityType: Optional[str] = "review"


class SentimentResult(BaseModel):
    text: str
    label: SentimentLabel
    score: float = Field(..., ge=0.0, le=1.0)
    positiveWords: list[str]
    negativeWords: list[str]
    entityId: Optional[str] = None
    entityType: Optional[str] = None
    analyzedAt: datetime = Field(default_factory=lambda: datetime.now(timezone.utc))


class BatchAnalyzeRequest(BaseModel):
    texts: list[str] = Field(..., min_length=1, max_length=50)

    @field_validator("texts")
    @classmethod
    def validate_texts(cls, v: list[str]) -> list[str]:
        if len(v) > 50:
            raise ValueError("Maximum 50 texts per batch request")
        for text in v:
            if not text.strip():
                raise ValueError("Empty strings are not allowed in batch")
        return v


class AggregateStats(BaseModel):
    entityType: Optional[str] = None
    positive: int
    negative: int
    neutral: int
    total: int
    avgScore: float
