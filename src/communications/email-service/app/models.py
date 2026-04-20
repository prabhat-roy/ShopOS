from __future__ import annotations

from datetime import datetime
from typing import Optional

from pydantic import BaseModel, Field


class EmailMessage(BaseModel):
    """Inbound event payload consumed from Kafka topic email.send."""

    messageId: str = Field(..., description="Unique identifier for this email message")
    to: str = Field(..., description="Recipient email address")
    subject: str = Field(..., description="Email subject line")
    body: str = Field(..., description="Plain-text body content")
    htmlBody: Optional[str] = Field(None, description="Optional HTML body content")
    from_addr: str = Field("noreply@shopos.com", alias="from", description="Sender address")
    templateId: Optional[str] = Field(None, description="Optional template identifier")
    metadata: dict = Field(default_factory=dict, description="Arbitrary key-value metadata")

    model_config = {"populate_by_name": True}


class EmailRecord(BaseModel):
    """Persisted record of a processed email attempt."""

    messageId: str
    to: str
    subject: str
    status: str = Field(..., description="'delivered' or 'failed'")
    sentAt: datetime
    errorMessage: Optional[str] = None
