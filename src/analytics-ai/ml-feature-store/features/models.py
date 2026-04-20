from datetime import datetime
from enum import Enum
from typing import Any, Optional

from pydantic import BaseModel, Field


class FeatureType(str, Enum):
    NUMERIC = "NUMERIC"
    CATEGORICAL = "CATEGORICAL"
    BOOLEAN = "BOOLEAN"
    VECTOR = "VECTOR"
    TEXT = "TEXT"


class FeatureDefinition(BaseModel):
    name: str
    featureGroup: str
    type: FeatureType
    description: str = ""
    tags: list[str] = Field(default_factory=list)
    defaultValue: Optional[Any] = None


class FeatureValue(BaseModel):
    entityId: str
    featureName: str
    featureGroup: str
    value: Any
    version: int = 1
    computedAt: datetime


class GetFeaturesRequest(BaseModel):
    entityId: str
    featureNames: list[str]
    featureGroup: str


class FeatureVector(BaseModel):
    entityId: str
    features: dict[str, Any]
    missingFeatures: list[str]
    retrievedAt: datetime
