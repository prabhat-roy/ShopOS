"""Pydantic models for label-service request/response schemas."""

from datetime import datetime
from typing import Any, Optional
from pydantic import BaseModel, Field


class Address(BaseModel):
    name: str
    street1: str
    street2: Optional[str] = None
    city: str
    state: str
    postal_code: str = Field(alias="postalCode")
    country: str = Field(default="US")

    model_config = {"populate_by_name": True}


class LabelRequest(BaseModel):
    shipment_id: str = Field(alias="shipmentId")
    tracking_number: str = Field(alias="trackingNumber")
    carrier: str
    from_address: Address = Field(alias="fromAddress")
    to_address: Address = Field(alias="toAddress")
    weight: float
    dimensions: Optional[dict[str, Any]] = None
    label_format: str = Field(default="ZPL", alias="labelFormat")

    model_config = {"populate_by_name": True}


class LabelResponse(BaseModel):
    shipment_id: str = Field(alias="shipmentId")
    tracking_number: str = Field(alias="trackingNumber")
    label_format: str = Field(alias="labelFormat")
    label_data: str = Field(alias="labelData")
    barcode_data: str = Field(alias="barcodeData")
    generated_at: datetime = Field(alias="generatedAt")

    model_config = {"populate_by_name": True, "populate_by_alias": True}
