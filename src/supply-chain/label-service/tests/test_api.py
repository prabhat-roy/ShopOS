"""Integration tests for label-service API endpoints — 6 tests."""

import pytest
from fastapi.testclient import TestClient

from main import app

client = TestClient(app)

# ---------------------------------------------------------------------------
# Shared payload helpers
# ---------------------------------------------------------------------------

FROM_ADDRESS = {
    "name": "Acme Warehouse",
    "street1": "100 Industrial Blvd",
    "city": "Newark",
    "state": "NJ",
    "postalCode": "07102",
    "country": "US",
}

TO_ADDRESS = {
    "name": "Jane Doe",
    "street1": "456 Elm Street",
    "street2": "Apt 3B",
    "city": "Los Angeles",
    "state": "CA",
    "postalCode": "90001",
    "country": "US",
}

VALID_LABEL_PAYLOAD = {
    "shipmentId": "SHIP-001",
    "trackingNumber": "1Z9999W99999999999",
    "carrier": "UPS",
    "fromAddress": FROM_ADDRESS,
    "toAddress": TO_ADDRESS,
    "weight": 2.5,
    "labelFormat": "ZPL",
}


def _make_payload(tracking_number: str, label_format: str = "ZPL") -> dict:
    return {**VALID_LABEL_PAYLOAD, "trackingNumber": tracking_number, "labelFormat": label_format}


# ---------------------------------------------------------------------------
# Tests
# ---------------------------------------------------------------------------

def test_healthz_returns_200():
    """GET /healthz must return HTTP 200 with status: ok."""
    res = client.get("/healthz")
    assert res.status_code == 200
    assert res.json()["status"] == "ok"


def test_generate_label_returns_200():
    """POST /labels/generate with valid payload returns HTTP 200."""
    res = client.post("/labels/generate", json=VALID_LABEL_PAYLOAD)
    assert res.status_code == 200
    data = res.json()
    assert data["trackingNumber"] == VALID_LABEL_PAYLOAD["trackingNumber"]
    assert data["labelFormat"] == "ZPL"
    assert data["labelData"]
    assert data["barcodeData"]
    assert data["generatedAt"]


def test_generate_label_text_format():
    """POST /labels/generate with labelFormat=TEXT returns a text label."""
    payload = {**VALID_LABEL_PAYLOAD, "labelFormat": "TEXT"}
    res = client.post("/labels/generate", json=payload)
    assert res.status_code == 200
    data = res.json()
    assert data["labelFormat"] == "TEXT"
    assert VALID_LABEL_PAYLOAD["trackingNumber"] in data["labelData"]


def test_batch_generate_returns_200():
    """POST /labels/batch with two items returns a list of two responses."""
    payloads = [
        _make_payload("TRK-BATCH-001", "ZPL"),
        _make_payload("TRK-BATCH-002", "TEXT"),
    ]
    res = client.post("/labels/batch", json=payloads)
    assert res.status_code == 200
    data = res.json()
    assert isinstance(data, list)
    assert len(data) == 2
    assert data[0]["trackingNumber"] == "TRK-BATCH-001"
    assert data[1]["trackingNumber"] == "TRK-BATCH-002"


def test_generate_label_missing_required_field_returns_422():
    """POST /labels/generate without carrier returns HTTP 422."""
    payload = {k: v for k, v in VALID_LABEL_PAYLOAD.items() if k != "carrier"}
    res = client.post("/labels/generate", json=payload)
    assert res.status_code == 422


def test_generate_label_missing_to_address_returns_422():
    """POST /labels/generate without toAddress returns HTTP 422."""
    payload = {k: v for k, v in VALID_LABEL_PAYLOAD.items() if k != "toAddress"}
    res = client.post("/labels/generate", json=payload)
    assert res.status_code == 422
