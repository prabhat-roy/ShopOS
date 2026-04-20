from __future__ import annotations

from datetime import datetime
from unittest.mock import AsyncMock, MagicMock

import pytest
from fastapi import FastAPI
from fastapi.testclient import TestClient

from tracker.api import router
from tracker.models import (
    BatchTrackingResponse,
    EventStats,
    EventType,
    StoredEvent,
    TrackingResponse,
)
from tracker.publisher import KafkaEventPublisher
from tracker.service import EventTrackingService
from tracker.store import InMemoryEventStore


def _build_test_app(service: EventTrackingService) -> FastAPI:
    app = FastAPI()
    app.include_router(router)
    app.state.service = service
    return app


def _make_stored(event_id: str = "evt-123", session_id: str = "sess-1") -> StoredEvent:
    return StoredEvent(
        eventId=event_id,
        eventType=EventType.PAGE_VIEW,
        sessionId=session_id,
        userId="u1",
        data={"page": "/home"},
        receivedAt=datetime(2024, 6, 10, 12, 0, 0),
    )


@pytest.fixture
def store() -> InMemoryEventStore:
    return InMemoryEventStore()


@pytest.fixture
def publisher() -> MagicMock:
    pub = MagicMock(spec=KafkaEventPublisher)
    pub.publish = AsyncMock(return_value=True)
    pub.batch_publish = AsyncMock(return_value=[True])
    return pub


@pytest.fixture
def service(store: InMemoryEventStore, publisher: MagicMock) -> EventTrackingService:
    return EventTrackingService(store=store, publisher=publisher)


@pytest.fixture
def client(service: EventTrackingService) -> TestClient:
    app = _build_test_app(service)
    return TestClient(app)


def test_healthz(client: TestClient) -> None:
    response = client.get("/healthz")
    assert response.status_code == 200
    body = response.json()
    assert body["status"] == "ok"
    assert body["service"] == "event-tracking-service"


def test_track_single_event_returns_202(client: TestClient) -> None:
    payload = {
        "eventType": "PAGE_VIEW",
        "sessionId": "sess-abc",
        "userId": "u1",
        "data": {"page": "/checkout"},
    }
    response = client.post("/track", json=payload)
    assert response.status_code == 202
    body = response.json()
    assert body["accepted"] is True
    assert "eventId" in body


def test_track_event_stored_and_retrievable(client: TestClient) -> None:
    payload = {
        "eventType": "CLICK",
        "sessionId": "sess-xyz",
        "data": {"productId": "prod-99"},
    }
    post_response = client.post("/track", json=payload)
    assert post_response.status_code == 202
    event_id = post_response.json()["eventId"]

    get_response = client.get(f"/events/{event_id}")
    assert get_response.status_code == 200
    body = get_response.json()
    assert body["eventId"] == event_id
    assert body["eventType"] == "CLICK"
    assert body["sessionId"] == "sess-xyz"


def test_track_batch_returns_202(client: TestClient) -> None:
    payload = {
        "events": [
            {"eventType": "PAGE_VIEW", "sessionId": "s1", "data": {}},
            {"eventType": "CLICK", "sessionId": "s1", "data": {"productId": "p1"}},
            {"eventType": "CONVERSION", "sessionId": "s2", "data": {"orderId": "o1"}},
        ]
    }
    response = client.post("/track/batch", json=payload)
    assert response.status_code == 202
    body = response.json()
    assert body["accepted"] == 3
    assert body["rejected"] == 0
    assert len(body["results"]) == 3


def test_batch_exceeding_100_events_rejected(client: TestClient) -> None:
    payload = {
        "events": [
            {"eventType": "PAGE_VIEW", "sessionId": "s1", "data": {}}
            for _ in range(101)
        ]
    }
    response = client.post("/track/batch", json=payload)
    assert response.status_code == 422


def test_get_nonexistent_event_returns_404(client: TestClient) -> None:
    response = client.get("/events/does-not-exist")
    assert response.status_code == 404


def test_list_events_by_session(client: TestClient) -> None:
    for i in range(3):
        client.post(
            "/track",
            json={
                "eventType": "PAGE_VIEW",
                "sessionId": "list-session",
                "data": {"index": i},
            },
        )
    client.post(
        "/track",
        json={"eventType": "PAGE_VIEW", "sessionId": "other-session", "data": {}},
    )

    response = client.get("/events?sessionId=list-session")
    assert response.status_code == 200
    body = response.json()
    assert len(body) == 3
    assert all(e["sessionId"] == "list-session" for e in body)


def test_get_stats(client: TestClient) -> None:
    client.post("/track", json={"eventType": "PAGE_VIEW", "sessionId": "s1", "data": {}})
    client.post("/track", json={"eventType": "IMPRESSION", "sessionId": "s2", "data": {}})

    response = client.get("/events/stats")
    assert response.status_code == 200
    body = response.json()
    assert "totalEvents" in body
    assert "byType" in body
    assert "uniqueSessions" in body
    assert body["totalEvents"] >= 2
