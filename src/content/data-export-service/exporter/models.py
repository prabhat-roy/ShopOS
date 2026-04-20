from __future__ import annotations

from datetime import datetime
from enum import Enum
from typing import Any, List, Optional

from pydantic import BaseModel, Field


class ExportFormat(str, Enum):
    CSV = "CSV"
    JSON = "JSON"
    XLSX = "XLSX"
    PDF = "PDF"


class ExportRequest(BaseModel):
    format: ExportFormat = Field(..., description="Target export format")
    filename: str = Field(..., min_length=1, description="Base filename (without extension)")
    headers: List[str] = Field(..., min_length=1, description="Column header names")
    rows: List[List[Any]] = Field(default_factory=list, description="Data rows; each inner list maps positionally to headers")
    title: Optional[str] = Field(None, description="Optional title displayed at the top of PDF exports")


class ExportResult(BaseModel):
    filename: str = Field(..., description="Final filename including extension")
    contentType: str = Field(..., description="MIME content type of the export")
    size: int = Field(..., description="Size of the export payload in bytes")
    rowCount: int = Field(..., description="Number of data rows exported")
    exportedAt: datetime = Field(..., description="UTC timestamp of the export")


class StreamExportRequest(BaseModel):
    """Identical to ExportRequest — kept separate for forward-compatibility with streaming."""
    format: ExportFormat = Field(..., description="Target export format")
    filename: str = Field(..., min_length=1, description="Base filename (without extension)")
    headers: List[str] = Field(..., min_length=1, description="Column header names")
    rows: List[List[Any]] = Field(default_factory=list, description="Data rows")
    title: Optional[str] = Field(None, description="Optional title for PDF exports")
