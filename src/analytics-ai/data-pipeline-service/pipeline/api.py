from typing import List, Optional

from fastapi import FastAPI, HTTPException, Query

from pipeline.models import EnrichedEvent, PipelineStats
from pipeline.store import PipelineStore

app = FastAPI(
    title="Data Pipeline Service",
    description="ETL pipeline: consumes raw analytics events, enriches them, and republishes.",
    version="1.0.0",
)

_store: Optional[PipelineStore] = None


def set_store(store: PipelineStore) -> None:
    global _store
    _store = store


def get_store() -> PipelineStore:
    if _store is None:
        raise RuntimeError("Store not initialised")
    return _store


@app.get("/healthz")
async def healthz():
    return {"status": "ok"}


@app.get("/pipeline/stats", response_model=PipelineStats)
async def pipeline_stats():
    return await get_store().get_stats()


@app.get("/pipeline/events", response_model=List[EnrichedEvent])
async def list_events(
    topic: Optional[str] = Query(default=None, description="Filter by Kafka topic"),
    limit: int = Query(default=50, ge=1, le=500, description="Max number of events to return"),
):
    return await get_store().list_events(topic=topic, limit=limit)


@app.get("/pipeline/events/{event_id}", response_model=EnrichedEvent)
async def get_event(event_id: str):
    event = await get_store().get_event(event_id)
    if event is None:
        raise HTTPException(status_code=404, detail=f"Event '{event_id}' not found")
    return event
