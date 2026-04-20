"""Unit tests for the ImageProcessor class."""

import base64
from io import BytesIO

import pytest
from PIL import Image

from processor.image_processor import ImageProcessor
from processor.models import ConvertRequest, ResizeRequest, ThumbnailRequest


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _make_image_bytes(width: int = 200, height: int = 100, fmt: str = "JPEG") -> bytes:
    """Create a simple solid-colour image in memory and return its bytes."""
    img = Image.new("RGB", (width, height), color=(100, 149, 237))
    buf = BytesIO()
    img.save(buf, format=fmt)
    return buf.getvalue()


def _make_rgba_image_bytes(width: int = 100, height: int = 100) -> bytes:
    img = Image.new("RGBA", (width, height), color=(100, 149, 237, 128))
    buf = BytesIO()
    img.save(buf, format="PNG")
    return buf.getvalue()


processor = ImageProcessor()


# ---------------------------------------------------------------------------
# Tests
# ---------------------------------------------------------------------------

class TestResize:
    def test_resize_maintains_aspect_ratio(self):
        """When maintainAspect=True a 200x100 image fitted into 100x100 yields 100x50."""
        img_bytes = _make_image_bytes(200, 100)
        req = ResizeRequest(width=100, height=100, maintainAspect=True, format="JPEG", quality=85)
        result = processor.resize(img_bytes, req)

        # Pillow thumbnail keeps within the box, preserving ratio
        assert result.width <= 100
        assert result.height <= 100
        # Confirm the aspect ratio is roughly preserved (width should be 2x height)
        ratio = result.width / result.height
        assert abs(ratio - 2.0) < 0.1

    def test_resize_ignores_aspect_ratio(self):
        """When maintainAspect=False the output must be exactly the requested size."""
        img_bytes = _make_image_bytes(200, 100)
        req = ResizeRequest(width=50, height=80, maintainAspect=False, format="JPEG", quality=85)
        result = processor.resize(img_bytes, req)

        assert result.width == 50
        assert result.height == 80

    def test_resize_returns_base64(self):
        """Result data field must be a valid base64 string."""
        img_bytes = _make_image_bytes(100, 100)
        req = ResizeRequest(width=50, height=50, maintainAspect=False, format="PNG", quality=85)
        result = processor.resize(img_bytes, req)

        decoded = base64.b64decode(result.data)
        img = Image.open(BytesIO(decoded))
        assert img.width == 50
        assert img.height == 50


class TestThumbnail:
    def test_thumbnail_is_square(self):
        """Thumbnail must always produce a square output image."""
        img_bytes = _make_image_bytes(300, 150)
        req = ThumbnailRequest(size=128)
        result = processor.thumbnail(img_bytes, req)

        assert result.width == result.height
        assert result.width == 128

    def test_thumbnail_default_size(self):
        img_bytes = _make_image_bytes(500, 200)
        req = ThumbnailRequest()
        result = processor.thumbnail(img_bytes, req)

        assert result.width == 256
        assert result.height == 256


class TestConvert:
    def test_convert_jpeg_to_png(self):
        """Converting a JPEG to PNG should produce a PNG result."""
        img_bytes = _make_image_bytes(100, 100, fmt="JPEG")
        req = ConvertRequest(format="PNG", quality=85)
        result = processor.convert(img_bytes, req)

        assert result.format == "PNG"
        decoded = base64.b64decode(result.data)
        img = Image.open(BytesIO(decoded))
        assert img.format == "PNG"

    def test_convert_to_webp(self):
        """Converting to WEBP should produce a WEBP image."""
        img_bytes = _make_image_bytes(80, 80, fmt="PNG")
        req = ConvertRequest(format="WEBP", quality=80)
        result = processor.convert(img_bytes, req)

        assert result.format == "WEBP"
        decoded = base64.b64decode(result.data)
        img = Image.open(BytesIO(decoded))
        assert img.format == "WEBP"

    def test_convert_invalid_format_raises_error(self):
        """Requesting an unsupported format should raise a ValueError via model validation."""
        with pytest.raises(ValueError, match="not supported"):
            ConvertRequest(format="TIFF", quality=85)


class TestOptimize:
    def test_optimize_returns_valid_image(self):
        """Optimized result must still be a valid image."""
        img_bytes = _make_image_bytes(200, 200, fmt="JPEG")
        result = processor.optimize(img_bytes, quality=60)

        decoded = base64.b64decode(result.data)
        img = Image.open(BytesIO(decoded))
        assert img.width == 200
        assert img.height == 200

    def test_optimize_size_recorded(self):
        """ProcessResult must record both original and processed sizes."""
        img_bytes = _make_image_bytes(300, 300, fmt="PNG")
        result = processor.optimize(img_bytes, quality=85)

        assert result.originalSize == len(img_bytes)
        assert result.processedSize > 0


class TestGetInfo:
    def test_get_info_returns_correct_dimensions(self):
        img_bytes = _make_image_bytes(320, 240, fmt="PNG")
        info = processor.get_info(img_bytes)

        assert info["width"] == 320
        assert info["height"] == 240
        assert info["format"] == "PNG"
        assert info["size"] == len(img_bytes)

    def test_get_info_includes_mode(self):
        img_bytes = _make_image_bytes(50, 50, fmt="JPEG")
        info = processor.get_info(img_bytes)

        assert "mode" in info
        assert info["mode"] == "RGB"


class TestOversizeHandling:
    def test_oversized_image_raises_value_error(self, monkeypatch):
        """Images exceeding MAX_FILE_SIZE_MB must raise ValueError."""
        from processor import config as cfg
        monkeypatch.setattr(cfg.settings, "MAX_FILE_SIZE_MB", 0)  # 0 MB → everything too large

        img_bytes = _make_image_bytes(10, 10, fmt="JPEG")
        with pytest.raises(ValueError, match="exceeds limit"):
            processor.get_info(img_bytes)


class TestQualityParameter:
    def test_quality_affects_jpeg_size(self):
        """A lower quality JPEG should be smaller than a higher quality one."""
        img_bytes = _make_image_bytes(400, 400, fmt="JPEG")
        req_high = ResizeRequest(width=400, height=400, maintainAspect=False, format="JPEG", quality=95)
        req_low = ResizeRequest(width=400, height=400, maintainAspect=False, format="JPEG", quality=10)

        result_high = processor.resize(img_bytes, req_high)
        result_low = processor.resize(img_bytes, req_low)

        assert result_low.processedSize < result_high.processedSize
