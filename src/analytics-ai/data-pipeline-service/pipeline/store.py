import asyncio
from abc import ABC, abstractmethod
from collections import defaultdict
from typing import Dict, List, Optional

from pipeline.models import EnrichedEvent, PipelineStats


class PipelineStore(ABC):
    @abstractmethod
    async def save_event(self, event: EnrichedEvent) -> None: ...

    @abstractmethod
    async def get_event(self, event_id: str) -> Optional[EnrichedEvent]: ...

    @abstractmethod
    async def list_events(self, topic: Optional[str], limit: int) -> List[EnrichedEvent]: ...

    @abstractmethod
    async def get_stats(self) -> PipelineStats: ...

    @abstractmethod
    async def increment_processed(self) -> None: ...

    @abstractmethod
    async def increment_failed(self) -> None: ...


class InMemoryPipelineStore(PipelineStore):
    def __init__(self) -> None:
        self._events: Dict[str, EnrichedEvent] = {}
        self._by_topic: Dict[str, List[str]] = defaultdict(list)
        self._lock = asyncio.Lock()
        self._processed = 0
        self._enriched = 0
        self._failed = 0
        self._total_ms = 0.0

    async def save_event(self, event: EnrichedEvent) -> None:
        async with self._lock:
            self._events[event.eventId] = event
            self._by_topic[event.topic].append(event.eventId)
            self._enriched += 1
            self._total_ms += event.processingTimeMs

    async def get_event(self, event_id: str) -> Optional[EnrichedEvent]:
        return self._events.get(event_id)

    async def list_events(self, topic: Optional[str], limit: int) -> List[EnrichedEvent]:
        if topic:
            ids = self._by_topic.get(topic, [])
            events = [self._events[eid] for eid in ids if eid in self._events]
        else:
            events = list(self._events.values())
        return sorted(events, key=lambda e: e.transformedAt, reverse=True)[:limit]

    async def get_stats(self) -> PipelineStats:
        avg_ms = (self._total_ms / self._enriched) if self._enriched > 0 else 0.0
        return PipelineStats(
            processed=self._processed,
            enriched=self._enriched,
            failed=self._failed,
            avgProcessingMs=round(avg_ms, 3),
        )

    async def increment_processed(self) -> None:
        async with self._lock:
            self._processed += 1

    async def increment_failed(self) -> None:
        async with self._lock:
            self._failed += 1


class CassandraPipelineStore(PipelineStore):
    """
    Production Cassandra-backed store.
    Falls back gracefully to an in-memory store when Cassandra is unavailable
    (useful in integration tests / local development without Cassandra).
    """

    def __init__(self, hosts: List[str], keyspace: str) -> None:
        self._hosts = hosts
        self._keyspace = keyspace
        self._session = None
        self._fallback = InMemoryPipelineStore()
        self._connected = False

    async def connect(self) -> None:
        try:
            from cassandra.cluster import Cluster  # type: ignore
            from cassandra.auth import PlainTextAuthProvider  # type: ignore
            import asyncio

            loop = asyncio.get_event_loop()
            cluster = Cluster(self._hosts)
            self._session = await loop.run_in_executor(None, cluster.connect, self._keyspace)
            self._connected = True
        except Exception:
            self._connected = False

    async def save_event(self, event: EnrichedEvent) -> None:
        if not self._connected:
            return await self._fallback.save_event(event)
        try:
            import json
            from cassandra.query import SimpleStatement  # type: ignore
            import asyncio

            cql = """
                INSERT INTO enriched_events (event_id, topic, original_data, enriched_data, transformed_at, processing_time_ms)
                VALUES (%s, %s, %s, %s, %s, %s)
            """
            loop = asyncio.get_event_loop()
            await loop.run_in_executor(
                None,
                self._session.execute,
                cql,
                (
                    event.eventId,
                    event.topic,
                    json.dumps(event.originalData),
                    json.dumps(event.enrichedData),
                    event.transformedAt,
                    event.processingTimeMs,
                ),
            )
        except Exception:
            await self._fallback.save_event(event)

    async def get_event(self, event_id: str) -> Optional[EnrichedEvent]:
        return await self._fallback.get_event(event_id)

    async def list_events(self, topic: Optional[str], limit: int) -> List[EnrichedEvent]:
        return await self._fallback.list_events(topic, limit)

    async def get_stats(self) -> PipelineStats:
        return await self._fallback.get_stats()

    async def increment_processed(self) -> None:
        await self._fallback.increment_processed()

    async def increment_failed(self) -> None:
        await self._fallback.increment_failed()
