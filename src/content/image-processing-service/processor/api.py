import json
from typing import Any, Dict

from fastapi import APIRouter, File, Form, HTTPException, UploadFile, status
from fastapi.responses import JSONResponse

from processor.config import settings
from processor.image_processor import ImageProcessor
from processor.models import ConvertRequest, ProcessResult, ResizeRequest, ThumbnailRequest

router = APIRouter()
_processor = ImageProcessor()

_MAX_BYTES = settings.MAX_FILE_SIZE_MB * 1024 * 1024


async def _read_upload(file: UploadFile) -> bytes:
    """Read uploaded file bytes and enforce size limit."""
    data = await file.read()
    if len(data) > _MAX_BYTES:
        raise HTTPException(
            status_code=status.HTTP_413_REQUEST_ENTITY_TOO_LARGE,
            detail=f"File size {len(data)} bytes exceeds maximum of {_MAX_BYTES} bytes "
                   f"({settings.MAX_FILE_SIZE_MB} MB).",
        )
    return data


def _validate_content_type(file: UploadFile) -> None:
    """Reject obviously non-image content types."""
    ct = (file.content_type or "").lower()
    if ct and not ct.startswith("image/"):
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Unsupported content type '{file.content_type}'. Expected an image.",
        )


@router.get("/healthz", tags=["health"])
async def healthz() -> Dict[str, str]:
    return {"status": "ok"}


@router.post(
    "/images/resize",
    response_model=ProcessResult,
    tags=["images"],
    summary="Resize an uploaded image",
)
async def resize_image(
    file: UploadFile = File(..., description="Image file to resize"),
    width: int = Form(..., gt=0, le=10000),
    height: int = Form(..., gt=0, le=10000),
    maintainAspect: bool = Form(True),
    format: str = Form("JPEG"),
    quality: int = Form(85, ge=1, le=100),
) -> ProcessResult:
    _validate_content_type(file)
    image_bytes = await _read_upload(file)
    try:
        req = ResizeRequest(
            width=width,
            height=height,
            maintainAspect=maintainAspect,
            format=format,
            quality=quality,
        )
        return _processor.resize(image_bytes, req)
    except ValueError as exc:
        raise HTTPException(status_code=status.HTTP_422_UNPROCESSABLE_ENTITY, detail=str(exc))


@router.post(
    "/images/convert",
    response_model=ProcessResult,
    tags=["images"],
    summary="Convert an uploaded image to a different format",
)
async def convert_image(
    file: UploadFile = File(..., description="Image file to convert"),
    format: str = Form(..., description="Target format: JPEG, PNG, WEBP, GIF, BMP"),
    quality: int = Form(85, ge=1, le=100),
) -> ProcessResult:
    _validate_content_type(file)
    image_bytes = await _read_upload(file)
    try:
        req = ConvertRequest(format=format, quality=quality)
        return _processor.convert(image_bytes, req)
    except ValueError as exc:
        raise HTTPException(status_code=status.HTTP_422_UNPROCESSABLE_ENTITY, detail=str(exc))


@router.post(
    "/images/thumbnail",
    response_model=ProcessResult,
    tags=["images"],
    summary="Create a square thumbnail from an uploaded image",
)
async def thumbnail_image(
    file: UploadFile = File(..., description="Image file to thumbnail"),
    size: int = Form(256, gt=0, le=4096),
) -> ProcessResult:
    _validate_content_type(file)
    image_bytes = await _read_upload(file)
    try:
        req = ThumbnailRequest(size=size)
        return _processor.thumbnail(image_bytes, req)
    except ValueError as exc:
        raise HTTPException(status_code=status.HTTP_422_UNPROCESSABLE_ENTITY, detail=str(exc))


@router.post(
    "/images/optimize",
    response_model=ProcessResult,
    tags=["images"],
    summary="Re-compress an image with optimisation flags",
)
async def optimize_image(
    file: UploadFile = File(..., description="Image file to optimize"),
    quality: int = Form(85, ge=1, le=100),
) -> ProcessResult:
    _validate_content_type(file)
    image_bytes = await _read_upload(file)
    try:
        return _processor.optimize(image_bytes, quality=quality)
    except ValueError as exc:
        raise HTTPException(status_code=status.HTTP_422_UNPROCESSABLE_ENTITY, detail=str(exc))


@router.post(
    "/images/info",
    tags=["images"],
    summary="Return metadata for an uploaded image without processing it",
)
async def image_info(
    file: UploadFile = File(..., description="Image file to inspect"),
) -> Dict[str, Any]:
    _validate_content_type(file)
    image_bytes = await _read_upload(file)
    try:
        return _processor.get_info(image_bytes)
    except ValueError as exc:
        raise HTTPException(status_code=status.HTTP_422_UNPROCESSABLE_ENTITY, detail=str(exc))
