from __future__ import annotations

from datetime import datetime
from enum import Enum
from typing import Any, Dict, List, Optional

from pydantic import BaseModel, Field, field_validator


class EventType(str, Enum):
    PAGE_VIEW = "PAGE_VIEW"
    CLICK = "CLICK"
    IMPRESSION = "IMPRESSION"
    CONVERSION = "CONVERSION"
    CUSTOM = "CUSTOM"


class TrackingEvent(BaseModel):
    eventType: EventType
    sessionId: str
    userId: Optional[str] = None
    data: Dict[str, Any] = Field(default_factory=dict)
    clientTimestamp: Optional[datetime] = None
    receivedAt: datetime = Field(default_factory=datetime.utcnow)
    ip: Optional[str] = None


class BatchTrackingRequest(BaseModel):
    events: List[TrackingEvent]

    @field_validator("events")
    @classmethod
    def validate_max_events(cls, v: List[TrackingEvent]) -> List[TrackingEvent]:
        if len(v) > 100:
            raise ValueError("Batch size must not exceed 100 events.")
        return v


class TrackingResponse(BaseModel):
    eventId: str
    accepted: bool
    message: str


class BatchTrackingResponse(BaseModel):
    accepted: int
    rejected: int
    results: List[TrackingResponse]


class StoredEvent(BaseModel):
    eventId: str
    eventType: EventType
    sessionId: str
    userId: Optional[str] = None
    data: Dict[str, Any] = Field(default_factory=dict)
    clientTimestamp: Optional[datetime] = None
    receivedAt: datetime
    ip: Optional[str] = None


class EventStats(BaseModel):
    totalEvents: int
    byType: Dict[str, int]
    uniqueSessions: int
