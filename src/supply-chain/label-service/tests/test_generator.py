"""Unit tests for LabelGenerator — 8 tests."""

import base64
from datetime import datetime, timezone, timedelta

import pytest

from label.generator import LabelGenerator
from label.models import Address, LabelRequest


# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------

@pytest.fixture
def from_address() -> Address:
    return Address(
        name="Acme Warehouse",
        street1="100 Industrial Blvd",
        city="Newark",
        state="NJ",
        postalCode="07102",
        country="US",
    )


@pytest.fixture
def to_address() -> Address:
    return Address(
        name="Jane Doe",
        street1="456 Elm Street",
        street2="Apt 3B",
        city="Los Angeles",
        state="CA",
        postalCode="90001",
        country="US",
    )


@pytest.fixture
def zpl_request(from_address, to_address) -> LabelRequest:
    return LabelRequest(
        shipmentId="SHIP-001",
        trackingNumber="1Z9999W99999999999",
        carrier="UPS",
        fromAddress=from_address,
        toAddress=to_address,
        weight=2.5,
        dimensions={"length": 12, "width": 8, "height": 4},
        labelFormat="ZPL",
    )


@pytest.fixture
def text_request(from_address, to_address) -> LabelRequest:
    return LabelRequest(
        shipmentId="SHIP-002",
        trackingNumber="1Z9999W99999999999",
        carrier="UPS",
        fromAddress=from_address,
        toAddress=to_address,
        weight=1.0,
        labelFormat="TEXT",
    )


@pytest.fixture
def generator() -> LabelGenerator:
    return LabelGenerator()


# ---------------------------------------------------------------------------
# Tests
# ---------------------------------------------------------------------------

def test_generate_zpl_contains_tracking_number(generator, zpl_request):
    """ZPL output must include the tracking number in a barcode field."""
    zpl = generator.generate_zpl(zpl_request)
    assert zpl_request.tracking_number in zpl


def test_generate_zpl_starts_with_xA_ends_with_xZ(generator, zpl_request):
    """ZPL output must be wrapped in ^XA ... ^XZ markers."""
    zpl = generator.generate_zpl(zpl_request)
    assert zpl.startswith("^XA")
    assert zpl.rstrip().endswith("^XZ")


def test_generate_text_contains_from_address(generator, text_request):
    """Text label must include the sender's name and street."""
    label = generator.generate_text(text_request)
    assert text_request.from_address.name in label
    assert text_request.from_address.street1 in label


def test_generate_text_contains_to_address(generator, text_request):
    """Text label must include the recipient's name and city."""
    label = generator.generate_text(text_request)
    assert text_request.to_address.name in label
    assert text_request.to_address.city in label


def test_barcode_data_is_non_empty(generator, zpl_request):
    """generate() must produce a non-empty barcodeData string."""
    response = generator.generate(zpl_request)
    assert response.barcode_data
    assert len(response.barcode_data) > 0


def test_barcode_data_is_valid_base64(generator, zpl_request):
    """barcodeData must be valid base64-encoded content."""
    response = generator.generate(zpl_request)
    # Should not raise
    decoded = base64.b64decode(response.barcode_data)
    assert len(decoded) > 0


def test_generate_text_format(generator, text_request):
    """Requesting TEXT format returns a LabelResponse with labelFormat TEXT."""
    response = generator.generate(text_request)
    assert response.label_format == "TEXT"
    assert text_request.tracking_number in response.label_data


def test_generated_at_is_recent(generator, zpl_request):
    """generatedAt must be within 5 seconds of now."""
    before = datetime.now(tz=timezone.utc)
    response = generator.generate(zpl_request)
    after = datetime.now(tz=timezone.utc)
    assert before - timedelta(seconds=1) <= response.generated_at <= after + timedelta(seconds=1)
