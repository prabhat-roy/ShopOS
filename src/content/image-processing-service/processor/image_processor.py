import base64
import time
from io import BytesIO
from typing import Dict, Any

from PIL import Image, UnidentifiedImageError

from processor.config import settings
from processor.models import (
    ConvertRequest,
    ProcessResult,
    ResizeRequest,
    ThumbnailRequest,
)

# Pillow format name normalization: "JPEG" is the canonical key used by Pillow
_FORMAT_SAVE_MAP: Dict[str, str] = {
    "JPEG": "JPEG",
    "PNG": "PNG",
    "WEBP": "WEBP",
    "GIF": "GIF",
    "BMP": "BMP",
}

# MIME types for reference (used by API layer)
FORMAT_MIME: Dict[str, str] = {
    "JPEG": "image/jpeg",
    "PNG": "image/png",
    "WEBP": "image/webp",
    "GIF": "image/gif",
    "BMP": "image/bmp",
}


def _open_image(image_bytes: bytes) -> Image.Image:
    """Open image bytes as a Pillow Image, raising ValueError on failure."""
    max_bytes = settings.MAX_FILE_SIZE_MB * 1024 * 1024
    if len(image_bytes) > max_bytes:
        raise ValueError(
            f"Image size {len(image_bytes)} bytes exceeds limit of {max_bytes} bytes "
            f"({settings.MAX_FILE_SIZE_MB} MB)."
        )
    try:
        img = Image.open(BytesIO(image_bytes))
        img.load()  # force decode so errors surface here
        return img
    except UnidentifiedImageError as exc:
        raise ValueError(f"Cannot identify image format: {exc}") from exc


def _save_image(img: Image.Image, fmt: str, quality: int) -> bytes:
    """Save a Pillow image to bytes in the given format."""
    buf = BytesIO()
    save_fmt = _FORMAT_SAVE_MAP[fmt]
    kwargs: Dict[str, Any] = {}

    if save_fmt == "JPEG":
        # JPEG does not support transparency — flatten to white background
        if img.mode in ("RGBA", "LA", "P"):
            background = Image.new("RGB", img.size, (255, 255, 255))
            if img.mode == "P":
                img = img.convert("RGBA")
            background.paste(img, mask=img.split()[-1] if img.mode in ("RGBA", "LA") else None)
            img = background
        elif img.mode != "RGB":
            img = img.convert("RGB")
        kwargs["quality"] = quality
        kwargs["optimize"] = True
    elif save_fmt == "PNG":
        kwargs["optimize"] = True
    elif save_fmt == "WEBP":
        kwargs["quality"] = quality
        kwargs["method"] = 6
    elif save_fmt == "GIF":
        if img.mode not in ("P", "L"):
            img = img.convert("P", palette=Image.Palette.ADAPTIVE)
    # BMP has no quality parameter

    img.save(buf, format=save_fmt, **kwargs)
    return buf.getvalue()


def _encode_result(
    original_size: int,
    processed_bytes: bytes,
    img: Image.Image,
    fmt: str,
    elapsed_ms: float,
) -> ProcessResult:
    return ProcessResult(
        originalSize=original_size,
        processedSize=len(processed_bytes),
        width=img.width,
        height=img.height,
        format=fmt,
        processingTimeMs=round(elapsed_ms, 3),
        data=base64.b64encode(processed_bytes).decode("utf-8"),
    )


class ImageProcessor:
    """Stateless image processing operations backed by Pillow."""

    def resize(self, image_bytes: bytes, req: ResizeRequest) -> ProcessResult:
        """Resize an image to the requested dimensions."""
        start = time.perf_counter()
        original_size = len(image_bytes)
        img = _open_image(image_bytes)

        if req.maintainAspect:
            img.thumbnail((req.width, req.height), Image.Resampling.LANCZOS)
        else:
            img = img.resize((req.width, req.height), Image.Resampling.LANCZOS)

        processed_bytes = _save_image(img, req.format, req.quality)
        elapsed_ms = (time.perf_counter() - start) * 1000
        return _encode_result(original_size, processed_bytes, img, req.format, elapsed_ms)

    def convert(self, image_bytes: bytes, req: ConvertRequest) -> ProcessResult:
        """Convert an image to a different format."""
        start = time.perf_counter()
        original_size = len(image_bytes)
        img = _open_image(image_bytes)

        processed_bytes = _save_image(img, req.format, req.quality)

        # Re-open processed bytes to get final dimensions (may differ for GIF/BMP)
        result_img = Image.open(BytesIO(processed_bytes))
        elapsed_ms = (time.perf_counter() - start) * 1000
        return _encode_result(original_size, processed_bytes, result_img, req.format, elapsed_ms)

    def thumbnail(self, image_bytes: bytes, req: ThumbnailRequest) -> ProcessResult:
        """Create a square thumbnail by cropping to center then resizing."""
        start = time.perf_counter()
        original_size = len(image_bytes)
        img = _open_image(image_bytes)

        # Crop to square using the shorter edge
        w, h = img.size
        min_side = min(w, h)
        left = (w - min_side) // 2
        top = (h - min_side) // 2
        img = img.crop((left, top, left + min_side, top + min_side))
        img = img.resize((req.size, req.size), Image.Resampling.LANCZOS)

        processed_bytes = _save_image(img, "JPEG", 85)
        elapsed_ms = (time.perf_counter() - start) * 1000
        return _encode_result(original_size, processed_bytes, img, "JPEG", elapsed_ms)

    def get_info(self, image_bytes: bytes) -> Dict[str, Any]:
        """Return metadata about an image without modifying it."""
        img = _open_image(image_bytes)
        fmt = img.format or "UNKNOWN"
        # Normalise PIL format name to our canonical names
        fmt = fmt.upper()
        return {
            "width": img.width,
            "height": img.height,
            "format": fmt,
            "mode": img.mode,
            "size": len(image_bytes),
        }

    def optimize(self, image_bytes: bytes, quality: int = 85) -> ProcessResult:
        """Re-save an image with optimisation flags to reduce file size."""
        start = time.perf_counter()
        original_size = len(image_bytes)
        img = _open_image(image_bytes)

        # Preserve original format if possible; fall back to JPEG
        raw_fmt = (img.format or "JPEG").upper()
        fmt = raw_fmt if raw_fmt in _FORMAT_SAVE_MAP else "JPEG"

        processed_bytes = _save_image(img, fmt, quality)

        # Re-open to confirm dimensions after optimisation
        result_img = Image.open(BytesIO(processed_bytes))
        elapsed_ms = (time.perf_counter() - start) * 1000
        return _encode_result(original_size, processed_bytes, result_img, fmt, elapsed_ms)
