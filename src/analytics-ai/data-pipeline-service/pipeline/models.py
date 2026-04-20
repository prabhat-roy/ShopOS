from datetime import datetime
from typing import Any, Dict, List, Optional
from pydantic import BaseModel, Field


class RawEvent(BaseModel):
    topic: str
    eventId: Optional[str] = None
    data: Dict[str, Any]
    receivedAt: datetime = Field(default_factory=datetime.utcnow)


class EnrichedEvent(BaseModel):
    eventId: str
    topic: str
    originalData: Dict[str, Any]
    enrichedData: Dict[str, Any]
    transformedAt: datetime = Field(default_factory=datetime.utcnow)
    processingTimeMs: float


class PipelineStats(BaseModel):
    processed: int
    enriched: int
    failed: int
    avgProcessingMs: float


class TransformRule(BaseModel):
    field: str
    action: str  # rename | drop | uppercase | lowercase | default
    params: Dict[str, Any] = Field(default_factory=dict)
