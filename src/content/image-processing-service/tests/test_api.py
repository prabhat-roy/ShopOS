"""Integration tests for the image-processing-service FastAPI endpoints."""

from io import BytesIO

import pytest
from fastapi.testclient import TestClient
from PIL import Image

from main import app

client = TestClient(app)


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _jpeg_bytes(width: int = 100, height: int = 100) -> bytes:
    img = Image.new("RGB", (width, height), color=(200, 100, 50))
    buf = BytesIO()
    img.save(buf, format="JPEG")
    return buf.getvalue()


def _png_bytes(width: int = 100, height: int = 100) -> bytes:
    img = Image.new("RGB", (width, height), color=(50, 150, 200))
    buf = BytesIO()
    img.save(buf, format="PNG")
    return buf.getvalue()


def _multipart_file(data: bytes, filename: str = "test.jpg", content_type: str = "image/jpeg"):
    return ("file", (filename, BytesIO(data), content_type))


# ---------------------------------------------------------------------------
# Tests
# ---------------------------------------------------------------------------

class TestHealthz:
    def test_healthz_returns_200(self):
        response = client.get("/healthz")
        assert response.status_code == 200
        assert response.json() == {"status": "ok"}


class TestResizeEndpoint:
    def test_resize_returns_200(self):
        response = client.post(
            "/images/resize",
            files=[_multipart_file(_jpeg_bytes(200, 100))],
            data={"width": "100", "height": "100", "maintainAspect": "true", "format": "JPEG", "quality": "85"},
        )
        assert response.status_code == 200
        body = response.json()
        assert "data" in body
        assert body["format"] == "JPEG"
        assert body["width"] <= 100
        assert body["height"] <= 100

    def test_resize_exact_dimensions(self):
        response = client.post(
            "/images/resize",
            files=[_multipart_file(_jpeg_bytes(200, 200))],
            data={"width": "50", "height": "80", "maintainAspect": "false", "format": "JPEG", "quality": "85"},
        )
        assert response.status_code == 200
        body = response.json()
        assert body["width"] == 50
        assert body["height"] == 80


class TestConvertEndpoint:
    def test_convert_jpeg_to_png(self):
        response = client.post(
            "/images/convert",
            files=[_multipart_file(_jpeg_bytes())],
            data={"format": "PNG", "quality": "85"},
        )
        assert response.status_code == 200
        assert response.json()["format"] == "PNG"

    def test_convert_to_webp(self):
        response = client.post(
            "/images/convert",
            files=[_multipart_file(_png_bytes())],
            data={"format": "WEBP", "quality": "80"},
        )
        assert response.status_code == 200
        assert response.json()["format"] == "WEBP"


class TestThumbnailEndpoint:
    def test_thumbnail_returns_square(self):
        response = client.post(
            "/images/thumbnail",
            files=[_multipart_file(_jpeg_bytes(300, 150))],
            data={"size": "128"},
        )
        assert response.status_code == 200
        body = response.json()
        assert body["width"] == 128
        assert body["height"] == 128


class TestOptimizeEndpoint:
    def test_optimize_returns_200(self):
        response = client.post(
            "/images/optimize",
            files=[_multipart_file(_jpeg_bytes(200, 200))],
            data={"quality": "70"},
        )
        assert response.status_code == 200
        body = response.json()
        assert body["processedSize"] > 0
        assert body["originalSize"] > 0


class TestInfoEndpoint:
    def test_info_returns_metadata(self):
        response = client.post(
            "/images/info",
            files=[_multipart_file(_jpeg_bytes(320, 240))],
        )
        assert response.status_code == 200
        body = response.json()
        assert body["width"] == 320
        assert body["height"] == 240
        assert "format" in body
        assert "mode" in body
        assert "size" in body


class TestValidationErrors:
    def test_invalid_file_type_returns_422(self):
        response = client.post(
            "/images/resize",
            files=[("file", ("bad.txt", BytesIO(b"not an image at all"), "text/plain"))],
            data={"width": "100", "height": "100"},
        )
        assert response.status_code == 422

    def test_missing_file_returns_422(self):
        response = client.post(
            "/images/resize",
            data={"width": "100", "height": "100"},
        )
        assert response.status_code == 422
