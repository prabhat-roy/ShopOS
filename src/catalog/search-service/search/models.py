from datetime import datetime
from typing import Any

from pydantic import BaseModel, Field


class ProductDocument(BaseModel):
    id: str
    sku: str
    name: str
    description: str
    category_id: str
    brand_id: str
    brand_name: str
    price: float
    currency: str = "USD"
    status: str = "active"
    tags: list[str] = Field(default_factory=list)
    attributes: dict[str, str] = Field(default_factory=dict)
    image_urls: list[str] = Field(default_factory=list)
    created_at: datetime


class SearchRequest(BaseModel):
    q: str = ""
    category_id: str = ""
    brand_id: str = ""
    min_price: float = 0
    max_price: float = 0
    tags: list[str] = Field(default_factory=list)
    status: str = "active"
    from_: int = Field(default=0, alias="from", ge=0)
    size: int = Field(default=20, ge=1, le=100)
    sort: str = "_score"

    model_config = {"populate_by_name": True}


class SearchResult(BaseModel):
    total: int
    hits: list[ProductDocument]
    aggregations: dict[str, Any] = Field(default_factory=dict)
    took_ms: int


class SuggestRequest(BaseModel):
    prefix: str
    size: int = 5
