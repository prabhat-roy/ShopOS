from __future__ import annotations

import logging
from typing import List, Optional

from fastapi import APIRouter, Depends, HTTPException, Query, Request

from tracker.models import (
    BatchTrackingRequest,
    BatchTrackingResponse,
    EventStats,
    StoredEvent,
    TrackingEvent,
    TrackingResponse,
)
from tracker.service import EventTrackingService

logger = logging.getLogger(__name__)
router = APIRouter()


def _get_service(request: Request) -> EventTrackingService:
    return request.app.state.service  # type: ignore[no-any-return]


@router.get("/healthz")
async def healthz() -> dict:
    return {"status": "ok", "service": "event-tracking-service"}


@router.post("/track", response_model=TrackingResponse, status_code=202)
async def track_event(
    event: TrackingEvent,
    request: Request,
    service: EventTrackingService = Depends(_get_service),
) -> TrackingResponse:
    if not event.ip:
        event = event.model_copy(update={"ip": request.client.host if request.client else None})
    return await service.track(event)


@router.post("/track/batch", response_model=BatchTrackingResponse, status_code=202)
async def track_batch(
    body: BatchTrackingRequest,
    request: Request,
    service: EventTrackingService = Depends(_get_service),
) -> BatchTrackingResponse:
    client_ip = request.client.host if request.client else None
    enriched = [
        event.model_copy(update={"ip": event.ip or client_ip})
        for event in body.events
    ]
    return await service.batch_track(enriched)


@router.get("/events/stats", response_model=EventStats)
async def get_stats(
    service: EventTrackingService = Depends(_get_service),
) -> EventStats:
    return await service.get_stats()


@router.get("/events/{event_id}", response_model=StoredEvent)
async def get_event(
    event_id: str,
    service: EventTrackingService = Depends(_get_service),
) -> StoredEvent:
    event = await service.get_event(event_id)
    if event is None:
        raise HTTPException(status_code=404, detail=f"Event '{event_id}' not found.")
    return event


@router.get("/events", response_model=List[StoredEvent])
async def list_events(
    sessionId: str = Query(..., description="Session ID to filter by"),
    limit: int = Query(default=50, ge=1, le=500),
    service: EventTrackingService = Depends(_get_service),
) -> List[StoredEvent]:
    return await service.list_session_events(sessionId, limit)
