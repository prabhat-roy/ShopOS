from pydantic import BaseModel, Field, field_validator
from typing import Optional


class ResizeRequest(BaseModel):
    width: int = Field(..., gt=0, le=10000, description="Target width in pixels")
    height: int = Field(..., gt=0, le=10000, description="Target height in pixels")
    maintainAspect: bool = Field(True, description="Maintain original aspect ratio")
    format: str = Field("JPEG", description="Output image format")
    quality: int = Field(85, ge=1, le=100, description="Output quality (1-100)")

    @field_validator("format")
    @classmethod
    def validate_format(cls, v: str) -> str:
        allowed = {"JPEG", "PNG", "WEBP", "GIF", "BMP"}
        upper = v.upper()
        if upper not in allowed:
            raise ValueError(f"Format '{v}' is not supported. Allowed: {allowed}")
        return upper


class ConvertRequest(BaseModel):
    format: str = Field(..., description="Target image format")
    quality: int = Field(85, ge=1, le=100, description="Output quality (1-100)")

    @field_validator("format")
    @classmethod
    def validate_format(cls, v: str) -> str:
        allowed = {"JPEG", "PNG", "WEBP", "GIF", "BMP"}
        upper = v.upper()
        if upper not in allowed:
            raise ValueError(f"Format '{v}' is not supported. Allowed: {allowed}")
        return upper


class ThumbnailRequest(BaseModel):
    size: int = Field(256, gt=0, le=4096, description="Square thumbnail side length in pixels")


class ProcessResult(BaseModel):
    originalSize: int = Field(..., description="Original image size in bytes")
    processedSize: int = Field(..., description="Processed image size in bytes")
    width: int = Field(..., description="Output image width in pixels")
    height: int = Field(..., description="Output image height in pixels")
    format: str = Field(..., description="Output image format")
    processingTimeMs: float = Field(..., description="Processing duration in milliseconds")
    data: str = Field(..., description="Base64-encoded processed image data")
