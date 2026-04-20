from __future__ import annotations

from datetime import datetime
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from tracker.models import EventType, TrackingEvent
from tracker.publisher import KafkaEventPublisher
from tracker.service import EventTrackingService
from tracker.store import InMemoryEventStore


def _make_event(
    event_type: EventType = EventType.PAGE_VIEW,
    session_id: str = "sess-1",
) -> TrackingEvent:
    return TrackingEvent(
        eventType=event_type,
        sessionId=session_id,
        userId="user-1",
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


@pytest.mark.asyncio
async def test_track_returns_accepted_response(
    service: EventTrackingService, publisher: MagicMock
) -> None:
    event = _make_event()
    response = await service.track(event)
    assert response.accepted is True
    assert response.eventId != ""
    assert "accepted" in response.message.lower()


@pytest.mark.asyncio
async def test_track_saves_event_to_store(
    service: EventTrackingService, store: InMemoryEventStore
) -> None:
    event = _make_event()
    response = await service.track(event)
    stored = await store.get_event(response.eventId)
    assert stored is not None
    assert stored.sessionId == event.sessionId
    assert stored.eventType == event.eventType


@pytest.mark.asyncio
async def test_track_calls_publisher(
    service: EventTrackingService, publisher: MagicMock
) -> None:
    event = _make_event()
    response = await service.track(event)
    publisher.publish.assert_awaited_once()
    call_args = publisher.publish.call_args
    assert call_args[0][0] == response.eventId
    assert call_args[0][1] == event


@pytest.mark.asyncio
async def test_track_returns_accepted_even_when_publish_fails(
    service: EventTrackingService, publisher: MagicMock
) -> None:
    publisher.publish = AsyncMock(return_value=False)
    event = _make_event()
    response = await service.track(event)
    assert response.accepted is True


@pytest.mark.asyncio
async def test_batch_track_all_accepted(
    service: EventTrackingService,
) -> None:
    events = [_make_event(session_id=f"sess-{i}") for i in range(5)]
    result = await service.batch_track(events)
    assert result.accepted == 5
    assert result.rejected == 0
    assert len(result.results) == 5


@pytest.mark.asyncio
async def test_get_event_by_id(
    service: EventTrackingService,
) -> None:
    event = _make_event()
    response = await service.track(event)
    stored = await service.get_event(response.eventId)
    assert stored is not None
    assert stored.eventId == response.eventId


@pytest.mark.asyncio
async def test_list_session_events_filtered_by_session(
    service: EventTrackingService,
) -> None:
    for _ in range(3):
        await service.track(_make_event(session_id="target-session"))
    await service.track(_make_event(session_id="other-session"))

    results = await service.list_session_events("target-session")
    assert len(results) == 3
    assert all(e.sessionId == "target-session" for e in results)


@pytest.mark.asyncio
async def test_get_stats_returns_correct_counts(
    service: EventTrackingService,
) -> None:
    await service.track(_make_event(EventType.PAGE_VIEW, "s1"))
    await service.track(_make_event(EventType.CLICK, "s2"))
    await service.track(_make_event(EventType.PAGE_VIEW, "s3"))

    stats = await service.get_stats()
    assert stats.totalEvents == 3
    assert stats.byType[EventType.PAGE_VIEW.value] == 2
    assert stats.byType[EventType.CLICK.value] == 1
    assert stats.uniqueSessions == 3
