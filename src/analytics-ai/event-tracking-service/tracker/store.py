from __future__ import annotations

import abc
import json
import logging
from typing import Dict, List, Optional

from tracker.models import EventStats, EventType, StoredEvent, TrackingEvent

logger = logging.getLogger(__name__)


class EventStore(abc.ABC):
    """Abstract base class for event persistence."""

    @abc.abstractmethod
    async def save_event(self, event_id: str, event: TrackingEvent) -> StoredEvent: ...

    @abc.abstractmethod
    async def get_event(self, event_id: str) -> Optional[StoredEvent]: ...

    @abc.abstractmethod
    async def list_events(
        self,
        session_id: str,
        limit: int = 50,
    ) -> List[StoredEvent]: ...

    @abc.abstractmethod
    async def get_stats(self) -> EventStats: ...


class InMemoryEventStore(EventStore):
    """In-memory implementation suitable for testing."""

    def __init__(self) -> None:
        self._events: Dict[str, StoredEvent] = {}

    async def save_event(self, event_id: str, event: TrackingEvent) -> StoredEvent:
        stored = StoredEvent(
            eventId=event_id,
            eventType=event.eventType,
            sessionId=event.sessionId,
            userId=event.userId,
            data=event.data,
            clientTimestamp=event.clientTimestamp,
            receivedAt=event.receivedAt,
            ip=event.ip,
        )
        self._events[event_id] = stored
        return stored

    async def get_event(self, event_id: str) -> Optional[StoredEvent]:
        return self._events.get(event_id)

    async def list_events(
        self,
        session_id: str,
        limit: int = 50,
    ) -> List[StoredEvent]:
        results = [e for e in self._events.values() if e.sessionId == session_id]
        results.sort(key=lambda e: e.receivedAt, reverse=True)
        return results[:limit]

    async def get_stats(self) -> EventStats:
        by_type: Dict[str, int] = {et.value: 0 for et in EventType}
        sessions: set = set()
        for event in self._events.values():
            by_type[event.eventType.value] += 1
            sessions.add(event.sessionId)
        return EventStats(
            totalEvents=len(self._events),
            byType=by_type,
            uniqueSessions=len(sessions),
        )


class CassandraEventStore(EventStore):
    """Cassandra-backed implementation for production use."""

    CREATE_KEYSPACE = """
        CREATE KEYSPACE IF NOT EXISTS {keyspace}
        WITH replication = {{'class': 'SimpleStrategy', 'replication_factor': '1'}}
    """

    CREATE_EVENTS_TABLE = """
        CREATE TABLE IF NOT EXISTS {keyspace}.events (
            event_id text PRIMARY KEY,
            event_type text,
            session_id text,
            user_id text,
            data text,
            client_timestamp timestamp,
            received_at timestamp,
            ip text
        )
    """

    CREATE_EVENTS_BY_SESSION_TABLE = """
        CREATE TABLE IF NOT EXISTS {keyspace}.events_by_session (
            session_id text,
            received_at timestamp,
            event_id text,
            event_type text,
            user_id text,
            data text,
            client_timestamp timestamp,
            ip text,
            PRIMARY KEY ((session_id), received_at, event_id)
        ) WITH CLUSTERING ORDER BY (received_at DESC)
    """

    def __init__(self, session, keyspace: str) -> None:  # type: ignore[no-untyped-def]
        self._session = session
        self._keyspace = keyspace
        self._init_schema()

    def _init_schema(self) -> None:
        self._session.execute(self.CREATE_KEYSPACE.format(keyspace=self._keyspace))
        self._session.execute(self.CREATE_EVENTS_TABLE.format(keyspace=self._keyspace))
        self._session.execute(self.CREATE_EVENTS_BY_SESSION_TABLE.format(keyspace=self._keyspace))
        logger.info("Cassandra schema initialised for keyspace '%s'", self._keyspace)

    async def save_event(self, event_id: str, event: TrackingEvent) -> StoredEvent:
        data_json = json.dumps(event.data)

        self._session.execute(
            f"INSERT INTO {self._keyspace}.events "  # noqa: S608
            f"(event_id, event_type, session_id, user_id, data, client_timestamp, received_at, ip) "
            f"VALUES (%s, %s, %s, %s, %s, %s, %s, %s)",
            (
                event_id,
                event.eventType.value,
                event.sessionId,
                event.userId,
                data_json,
                event.clientTimestamp,
                event.receivedAt,
                event.ip,
            ),
        )

        self._session.execute(
            f"INSERT INTO {self._keyspace}.events_by_session "  # noqa: S608
            f"(session_id, received_at, event_id, event_type, user_id, data, client_timestamp, ip) "
            f"VALUES (%s, %s, %s, %s, %s, %s, %s, %s)",
            (
                event.sessionId,
                event.receivedAt,
                event_id,
                event.eventType.value,
                event.userId,
                data_json,
                event.clientTimestamp,
                event.ip,
            ),
        )

        return StoredEvent(
            eventId=event_id,
            eventType=event.eventType,
            sessionId=event.sessionId,
            userId=event.userId,
            data=event.data,
            clientTimestamp=event.clientTimestamp,
            receivedAt=event.receivedAt,
            ip=event.ip,
        )

    async def get_event(self, event_id: str) -> Optional[StoredEvent]:
        rows = self._session.execute(
            f"SELECT * FROM {self._keyspace}.events WHERE event_id = %s",  # noqa: S608
            (event_id,),
        )
        row = rows.one()
        if row is None:
            return None
        return StoredEvent(
            eventId=row.event_id,
            eventType=EventType(row.event_type),
            sessionId=row.session_id,
            userId=row.user_id,
            data=json.loads(row.data) if row.data else {},
            clientTimestamp=row.client_timestamp,
            receivedAt=row.received_at,
            ip=row.ip,
        )

    async def list_events(
        self,
        session_id: str,
        limit: int = 50,
    ) -> List[StoredEvent]:
        rows = self._session.execute(
            f"SELECT * FROM {self._keyspace}.events_by_session "  # noqa: S608
            f"WHERE session_id = %s LIMIT %s",
            (session_id, limit),
        )
        return [
            StoredEvent(
                eventId=row.event_id,
                eventType=EventType(row.event_type),
                sessionId=row.session_id,
                userId=row.user_id,
                data=json.loads(row.data) if row.data else {},
                clientTimestamp=row.client_timestamp,
                receivedAt=row.received_at,
                ip=row.ip,
            )
            for row in rows
        ]

    async def get_stats(self) -> EventStats:
        rows = self._session.execute(
            f"SELECT event_type, session_id FROM {self._keyspace}.events"  # noqa: S608
        )
        by_type: Dict[str, int] = {et.value: 0 for et in EventType}
        sessions: set = set()
        total = 0
        for row in rows:
            total += 1
            by_type[row.event_type] = by_type.get(row.event_type, 0) + 1
            sessions.add(row.session_id)
        return EventStats(
            totalEvents=total,
            byType=by_type,
            uniqueSessions=len(sessions),
        )
