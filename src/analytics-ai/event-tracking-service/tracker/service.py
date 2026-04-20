from __future__ import annotations

import logging
import uuid
from typing import List, Optional

from tracker.models import (
    BatchTrackingResponse,
    EventStats,
    StoredEvent,
    TrackingEvent,
    TrackingResponse,
)
from tracker.publisher import KafkaEventPublisher
from tracker.store import EventStore

logger = logging.getLogger(__name__)


class EventTrackingService:
    """Orchestrates event ingestion: persists to store and publishes to Kafka."""

    def __init__(self, store: EventStore, publisher: KafkaEventPublisher) -> None:
        self._store = store
        self._publisher = publisher

    async def track(self, event: TrackingEvent) -> TrackingResponse:
        event_id = str(uuid.uuid4())
        try:
            await self._store.save_event(event_id, event)
        except Exception as exc:  # noqa: BLE001
            logger.error("Failed to persist event %s: %s", event_id, exc)
            return TrackingResponse(
                eventId=event_id,
                accepted=False,
                message=f"Storage error: {exc}",
            )

        published = await self._publisher.publish(event_id, event)
        if not published:
            logger.warning("Event %s stored but not published to Kafka.", event_id)

        return TrackingResponse(
            eventId=event_id,
            accepted=True,
            message="Event accepted.",
        )

    async def batch_track(self, events: List[TrackingEvent]) -> BatchTrackingResponse:
        results: List[TrackingResponse] = []
        for event in events:
            result = await self.track(event)
            results.append(result)

        accepted = sum(1 for r in results if r.accepted)
        rejected = len(results) - accepted

        return BatchTrackingResponse(
            accepted=accepted,
            rejected=rejected,
            results=results,
        )

    async def get_event(self, event_id: str) -> Optional[StoredEvent]:
        return await self._store.get_event(event_id)

    async def list_session_events(
        self,
        session_id: str,
        limit: int = 50,
    ) -> List[StoredEvent]:
        return await self._store.list_events(session_id, limit)

    async def get_stats(self) -> EventStats:
        return await self._store.get_stats()
