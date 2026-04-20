import pytest
import pytest_asyncio
from datetime import datetime, timezone
from httpx import AsyncClient, ASGITransport

from pipeline.api import app, set_store
from pipeline.models import EnrichedEvent
from pipeline.store import InMemoryPipelineStore


@pytest_asyncio.fixture
async def client():
    store = InMemoryPipelineStore()
    set_store(store)
    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://test") as ac:
        yield ac, store


@pytest.mark.asyncio
async def test_healthz_returns_ok(client):
    ac, _ = client
    response = await ac.get("/healthz")
    assert response.status_code == 200
    assert response.json()["status"] == "ok"


@pytest.mark.asyncio
async def test_pipeline_stats_empty(client):
    ac, _ = client
    response = await ac.get("/pipeline/stats")
    assert response.status_code == 200
    body = response.json()
    assert body["processed"] == 0
    assert body["enriched"] == 0
    assert body["failed"] == 0
    assert body["avgProcessingMs"] == 0.0


@pytest.mark.asyncio
async def test_list_events_empty(client):
    ac, _ = client
    response = await ac.get("/pipeline/events")
    assert response.status_code == 200
    assert response.json() == []


@pytest.mark.asyncio
async def test_get_event_not_found(client):
    ac, _ = client
    response = await ac.get("/pipeline/events/nonexistent-id")
    assert response.status_code == 404
    assert "not found" in response.json()["detail"].lower()


@pytest.mark.asyncio
async def test_list_events_after_save(client):
    ac, store = client
    event = EnrichedEvent(
        eventId="evt-abc-123",
        topic="analytics.page.viewed",
        originalData={"sessionId": "s-1"},
        enrichedData={"sessionId": "s-1", "platform": "windows"},
        transformedAt=datetime.now(timezone.utc),
        processingTimeMs=1.23,
    )
    await store.save_event(event)
    response = await ac.get("/pipeline/events")
    assert response.status_code == 200
    body = response.json()
    assert len(body) == 1
    assert body[0]["eventId"] == "evt-abc-123"


@pytest.mark.asyncio
async def test_get_event_by_id(client):
    ac, store = client
    event = EnrichedEvent(
        eventId="evt-xyz-999",
        topic="analytics.product.clicked",
        originalData={"productId": "p-1"},
        enrichedData={"productId": "p-1", "geo_region": "internal"},
        transformedAt=datetime.now(timezone.utc),
        processingTimeMs=0.5,
    )
    await store.save_event(event)
    response = await ac.get("/pipeline/events/evt-xyz-999")
    assert response.status_code == 200
    assert response.json()["eventId"] == "evt-xyz-999"
    assert response.json()["topic"] == "analytics.product.clicked"


@pytest.mark.asyncio
async def test_list_events_topic_filter(client):
    ac, store = client
    for i, topic in enumerate(["analytics.page.viewed", "analytics.product.clicked", "analytics.page.viewed"]):
        event = EnrichedEvent(
            eventId=f"evt-{i}",
            topic=topic,
            originalData={},
            enrichedData={},
            transformedAt=datetime.now(timezone.utc),
            processingTimeMs=0.1,
        )
        await store.save_event(event)

    response = await ac.get("/pipeline/events?topic=analytics.page.viewed")
    assert response.status_code == 200
    body = response.json()
    assert len(body) == 2
    assert all(e["topic"] == "analytics.page.viewed" for e in body)


@pytest.mark.asyncio
async def test_pipeline_stats_after_save(client):
    ac, store = client
    await store.increment_processed()
    await store.increment_processed()
    await store.increment_failed()
    event = EnrichedEvent(
        eventId="e-1",
        topic="analytics.page.viewed",
        originalData={},
        enrichedData={},
        transformedAt=datetime.now(timezone.utc),
        processingTimeMs=2.0,
    )
    await store.save_event(event)
    response = await ac.get("/pipeline/stats")
    assert response.status_code == 200
    body = response.json()
    assert body["processed"] == 2
    assert body["failed"] == 1
    assert body["enriched"] == 1
    assert body["avgProcessingMs"] == 2.0
