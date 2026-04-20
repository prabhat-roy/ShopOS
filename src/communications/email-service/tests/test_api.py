"""Tests for the FastAPI router (app/api.py)."""
from __future__ import annotations

from datetime import datetime, timezone
from unittest.mock import AsyncMock, patch

import pytest
from fastapi import FastAPI
from fastapi.testclient import TestClient

from app.api import router
from app.models import EmailRecord


# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------


@pytest.fixture()
def client():
    application = FastAPI()
    application.include_router(router)
    return TestClient(application)


def _record(message_id: str = "msg-001", status: str = "delivered") -> EmailRecord:
    return EmailRecord(
        messageId=message_id,
        to="user@example.com",
        subject="Test Subject",
        status=status,
        sentAt=datetime(2024, 1, 15, 12, 0, 0, tzinfo=timezone.utc),
    )


# ---------------------------------------------------------------------------
# /healthz
# ---------------------------------------------------------------------------


class TestHealthz:
    def test_returns_200(self, client):
        with patch("app.api.kafka_consumer") as mock_consumer:
            mock_consumer.is_running = True
            resp = client.get("/healthz")
        assert resp.status_code == 200

    def test_status_ok(self, client):
        with patch("app.api.kafka_consumer") as mock_consumer:
            mock_consumer.is_running = True
            resp = client.get("/healthz")
        assert resp.json()["status"] == "ok"

    def test_consumer_running_reflected(self, client):
        with patch("app.api.kafka_consumer") as mock_consumer:
            mock_consumer.is_running = True
            resp = client.get("/healthz")
        assert resp.json()["consumer"] == "running"

    def test_consumer_stopped_reflected(self, client):
        with patch("app.api.kafka_consumer") as mock_consumer:
            mock_consumer.is_running = False
            resp = client.get("/healthz")
        assert resp.json()["consumer"] == "stopped"


# ---------------------------------------------------------------------------
# GET /emails/{message_id}
# ---------------------------------------------------------------------------


class TestGetEmail:
    def test_found_returns_200(self, client):
        rec = _record()
        with patch("app.api.email_store") as mock_store:
            mock_store.get = AsyncMock(return_value=rec)
            resp = client.get("/emails/msg-001")
        assert resp.status_code == 200

    def test_found_returns_correct_body(self, client):
        rec = _record(message_id="msg-XYZ")
        with patch("app.api.email_store") as mock_store:
            mock_store.get = AsyncMock(return_value=rec)
            resp = client.get("/emails/msg-XYZ")
        data = resp.json()
        assert data["messageId"] == "msg-XYZ"
        assert data["status"] == "delivered"

    def test_not_found_returns_404(self, client):
        with patch("app.api.email_store") as mock_store:
            mock_store.get = AsyncMock(return_value=None)
            resp = client.get("/emails/missing-id")
        assert resp.status_code == 404


# ---------------------------------------------------------------------------
# GET /emails
# ---------------------------------------------------------------------------


class TestListEmails:
    def test_list_returns_200(self, client):
        records = [_record(f"msg-{i}") for i in range(3)]
        with patch("app.api.email_store") as mock_store:
            mock_store.list = AsyncMock(return_value=records)
            resp = client.get("/emails")
        assert resp.status_code == 200

    def test_list_returns_array(self, client):
        records = [_record("msg-1"), _record("msg-2")]
        with patch("app.api.email_store") as mock_store:
            mock_store.list = AsyncMock(return_value=records)
            resp = client.get("/emails")
        assert isinstance(resp.json(), list)
        assert len(resp.json()) == 2


# ---------------------------------------------------------------------------
# GET /email/stats
# ---------------------------------------------------------------------------


class TestGetStats:
    def test_stats_returns_200(self, client):
        with patch("app.api.email_store") as mock_store:
            mock_store.get_stats = AsyncMock(return_value={"sent": 10, "delivered": 9, "failed": 1})
            resp = client.get("/email/stats")
        assert resp.status_code == 200

    def test_stats_body_has_keys(self, client):
        stats = {"sent": 5, "delivered": 4, "failed": 1}
        with patch("app.api.email_store") as mock_store:
            mock_store.get_stats = AsyncMock(return_value=stats)
            resp = client.get("/email/stats")
        body = resp.json()
        assert "sent" in body
        assert "delivered" in body
        assert "failed" in body
