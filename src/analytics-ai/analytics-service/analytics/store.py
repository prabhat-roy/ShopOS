from __future__ import annotations

import abc
import logging
from collections import defaultdict
from datetime import datetime
from typing import List, Optional

from analytics.models import (
    AnalyticsSummary,
    PageViewEvent,
    PageViewRecord,
    ProductClickEvent,
    ProductClickRecord,
    SearchEvent,
    SearchRecord,
    SearchQueryStat,
    TopProduct,
)

logger = logging.getLogger(__name__)


class AnalyticsStore(abc.ABC):
    """Abstract base class for analytics persistence."""

    @abc.abstractmethod
    async def save_page_view(self, event: PageViewEvent) -> None: ...

    @abc.abstractmethod
    async def save_product_click(self, event: ProductClickEvent) -> None: ...

    @abc.abstractmethod
    async def save_search(self, event: SearchEvent) -> None: ...

    @abc.abstractmethod
    async def get_page_views(
        self,
        start_date: datetime,
        end_date: datetime,
        limit: int = 100,
    ) -> List[PageViewRecord]: ...

    @abc.abstractmethod
    async def get_top_products(
        self,
        start_date: datetime,
        end_date: datetime,
        limit: int = 100,
    ) -> List[TopProduct]: ...

    @abc.abstractmethod
    async def get_search_queries(
        self,
        start_date: datetime,
        end_date: datetime,
        limit: int = 100,
    ) -> List[SearchQueryStat]: ...

    @abc.abstractmethod
    async def get_summary(
        self,
        start_date: datetime,
        end_date: datetime,
    ) -> AnalyticsSummary: ...


class InMemoryAnalyticsStore(AnalyticsStore):
    """In-memory implementation suitable for testing."""

    def __init__(self) -> None:
        self._page_views: List[PageViewRecord] = []
        self._product_clicks: List[ProductClickRecord] = []
        self._searches: List[SearchRecord] = []

    async def save_page_view(self, event: PageViewEvent) -> None:
        record = PageViewRecord(
            sessionId=event.sessionId,
            userId=event.userId,
            pageUrl=event.pageUrl,
            referrer=event.referrer,
            userAgent=event.userAgent,
            timestamp=event.timestamp,
        )
        self._page_views.append(record)

    async def save_product_click(self, event: ProductClickEvent) -> None:
        record = ProductClickRecord(
            sessionId=event.sessionId,
            userId=event.userId,
            productId=event.productId,
            sku=event.sku,
            position=event.position,
            clickCount=1,
            timestamp=event.timestamp,
        )
        self._product_clicks.append(record)

    async def save_search(self, event: SearchEvent) -> None:
        record = SearchRecord(
            sessionId=event.sessionId,
            userId=event.userId,
            query=event.query,
            resultCount=event.resultCount,
            timestamp=event.timestamp,
        )
        self._searches.append(record)

    async def get_page_views(
        self,
        start_date: datetime,
        end_date: datetime,
        limit: int = 100,
    ) -> List[PageViewRecord]:
        results = [
            pv
            for pv in self._page_views
            if start_date <= pv.timestamp <= end_date
        ]
        results.sort(key=lambda r: r.timestamp, reverse=True)
        return results[:limit]

    async def get_top_products(
        self,
        start_date: datetime,
        end_date: datetime,
        limit: int = 100,
    ) -> List[TopProduct]:
        counts: dict[str, dict] = defaultdict(lambda: {"sku": "", "count": 0})
        for click in self._product_clicks:
            if start_date <= click.timestamp <= end_date:
                counts[click.productId]["sku"] = click.sku
                counts[click.productId]["count"] += 1

        top = sorted(counts.items(), key=lambda x: x[1]["count"], reverse=True)
        return [
            TopProduct(productId=pid, sku=data["sku"], clickCount=data["count"])
            for pid, data in top[:limit]
        ]

    async def get_search_queries(
        self,
        start_date: datetime,
        end_date: datetime,
        limit: int = 100,
    ) -> List[SearchQueryStat]:
        agg: dict[str, dict] = defaultdict(lambda: {"count": 0, "total_results": 0})
        for search in self._searches:
            if start_date <= search.timestamp <= end_date:
                agg[search.query]["count"] += 1
                agg[search.query]["total_results"] += search.resultCount

        stats = sorted(agg.items(), key=lambda x: x[1]["count"], reverse=True)
        return [
            SearchQueryStat(
                query=q,
                searchCount=data["count"],
                avgResultCount=data["total_results"] / data["count"],
            )
            for q, data in stats[:limit]
        ]

    async def get_summary(
        self,
        start_date: datetime,
        end_date: datetime,
    ) -> AnalyticsSummary:
        pv_in_range = [pv for pv in self._page_views if start_date <= pv.timestamp <= end_date]
        pc_in_range = [pc for pc in self._product_clicks if start_date <= pc.timestamp <= end_date]
        sr_in_range = [sr for sr in self._searches if start_date <= sr.timestamp <= end_date]

        all_sessions = (
            {pv.sessionId for pv in pv_in_range}
            | {pc.sessionId for pc in pc_in_range}
            | {sr.sessionId for sr in sr_in_range}
        )

        return AnalyticsSummary(
            totalPageViews=len(pv_in_range),
            totalProductClicks=len(pc_in_range),
            totalSearches=len(sr_in_range),
            uniqueSessions=len(all_sessions),
            startDate=start_date,
            endDate=end_date,
        )


class CassandraAnalyticsStore(AnalyticsStore):
    """Cassandra-backed implementation for production use."""

    CREATE_KEYSPACE = """
        CREATE KEYSPACE IF NOT EXISTS {keyspace}
        WITH replication = {{'class': 'SimpleStrategy', 'replication_factor': '1'}}
    """

    CREATE_PAGE_VIEWS_TABLE = """
        CREATE TABLE IF NOT EXISTS {keyspace}.page_views (
            date_bucket text,
            timestamp timestamp,
            session_id text,
            user_id text,
            page_url text,
            referrer text,
            user_agent text,
            PRIMARY KEY ((date_bucket), timestamp, session_id)
        ) WITH CLUSTERING ORDER BY (timestamp DESC)
    """

    CREATE_PRODUCT_CLICKS_TABLE = """
        CREATE TABLE IF NOT EXISTS {keyspace}.product_clicks (
            date_bucket text,
            timestamp timestamp,
            session_id text,
            user_id text,
            product_id text,
            sku text,
            position int,
            PRIMARY KEY ((date_bucket), timestamp, session_id, product_id)
        ) WITH CLUSTERING ORDER BY (timestamp DESC)
    """

    CREATE_SEARCH_EVENTS_TABLE = """
        CREATE TABLE IF NOT EXISTS {keyspace}.search_events (
            date_bucket text,
            timestamp timestamp,
            session_id text,
            user_id text,
            query text,
            result_count int,
            PRIMARY KEY ((date_bucket), timestamp, session_id)
        ) WITH CLUSTERING ORDER BY (timestamp DESC)
    """

    INSERT_PAGE_VIEW = """
        INSERT INTO {keyspace}.page_views
            (date_bucket, timestamp, session_id, user_id, page_url, referrer, user_agent)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    """

    INSERT_PRODUCT_CLICK = """
        INSERT INTO {keyspace}.product_clicks
            (date_bucket, timestamp, session_id, user_id, product_id, sku, position)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    """

    INSERT_SEARCH_EVENT = """
        INSERT INTO {keyspace}.search_events
            (date_bucket, timestamp, session_id, user_id, query, result_count)
        VALUES (?, ?, ?, ?, ?, ?)
    """

    def __init__(self, session, keyspace: str) -> None:  # type: ignore[no-untyped-def]
        self._session = session
        self._keyspace = keyspace
        self._init_schema()

    def _init_schema(self) -> None:
        self._session.execute(self.CREATE_KEYSPACE.format(keyspace=self._keyspace))
        self._session.execute(self.CREATE_PAGE_VIEWS_TABLE.format(keyspace=self._keyspace))
        self._session.execute(self.CREATE_PRODUCT_CLICKS_TABLE.format(keyspace=self._keyspace))
        self._session.execute(self.CREATE_SEARCH_EVENTS_TABLE.format(keyspace=self._keyspace))
        logger.info("Cassandra schema initialised for keyspace '%s'", self._keyspace)

    @staticmethod
    def _date_bucket(ts: datetime) -> str:
        return ts.strftime("%Y-%m-%d")

    @staticmethod
    def _date_buckets_between(start: datetime, end: datetime) -> List[str]:
        from datetime import timedelta

        buckets = []
        current = start.replace(hour=0, minute=0, second=0, microsecond=0)
        while current <= end:
            buckets.append(current.strftime("%Y-%m-%d"))
            current += timedelta(days=1)
        return buckets

    async def save_page_view(self, event: PageViewEvent) -> None:
        stmt = self._session.prepare(
            self.INSERT_PAGE_VIEW.format(keyspace=self._keyspace)
        )
        self._session.execute(
            stmt,
            (
                self._date_bucket(event.timestamp),
                event.timestamp,
                event.sessionId,
                event.userId,
                event.pageUrl,
                event.referrer,
                event.userAgent,
            ),
        )

    async def save_product_click(self, event: ProductClickEvent) -> None:
        stmt = self._session.prepare(
            self.INSERT_PRODUCT_CLICK.format(keyspace=self._keyspace)
        )
        self._session.execute(
            stmt,
            (
                self._date_bucket(event.timestamp),
                event.timestamp,
                event.sessionId,
                event.userId,
                event.productId,
                event.sku,
                event.position,
            ),
        )

    async def save_search(self, event: SearchEvent) -> None:
        stmt = self._session.prepare(
            self.INSERT_SEARCH_EVENT.format(keyspace=self._keyspace)
        )
        self._session.execute(
            stmt,
            (
                self._date_bucket(event.timestamp),
                event.timestamp,
                event.sessionId,
                event.userId,
                event.query,
                event.resultCount,
            ),
        )

    async def get_page_views(
        self,
        start_date: datetime,
        end_date: datetime,
        limit: int = 100,
    ) -> List[PageViewRecord]:
        buckets = self._date_buckets_between(start_date, end_date)
        results: List[PageViewRecord] = []
        for bucket in buckets:
            rows = self._session.execute(
                f"SELECT * FROM {self._keyspace}.page_views "  # noqa: S608
                f"WHERE date_bucket = %s AND timestamp >= %s AND timestamp <= %s "
                f"LIMIT %s ALLOW FILTERING",
                (bucket, start_date, end_date, limit),
            )
            for row in rows:
                results.append(
                    PageViewRecord(
                        sessionId=row.session_id,
                        userId=row.user_id,
                        pageUrl=row.page_url,
                        referrer=row.referrer,
                        userAgent=row.user_agent,
                        timestamp=row.timestamp,
                    )
                )
            if len(results) >= limit:
                break
        results.sort(key=lambda r: r.timestamp, reverse=True)
        return results[:limit]

    async def get_top_products(
        self,
        start_date: datetime,
        end_date: datetime,
        limit: int = 100,
    ) -> List[TopProduct]:
        buckets = self._date_buckets_between(start_date, end_date)
        counts: dict[str, dict] = defaultdict(lambda: {"sku": "", "count": 0})
        for bucket in buckets:
            rows = self._session.execute(
                f"SELECT product_id, sku FROM {self._keyspace}.product_clicks "  # noqa: S608
                f"WHERE date_bucket = %s AND timestamp >= %s AND timestamp <= %s "
                f"ALLOW FILTERING",
                (bucket, start_date, end_date),
            )
            for row in rows:
                counts[row.product_id]["sku"] = row.sku
                counts[row.product_id]["count"] += 1

        top = sorted(counts.items(), key=lambda x: x[1]["count"], reverse=True)
        return [
            TopProduct(productId=pid, sku=data["sku"], clickCount=data["count"])
            for pid, data in top[:limit]
        ]

    async def get_search_queries(
        self,
        start_date: datetime,
        end_date: datetime,
        limit: int = 100,
    ) -> List[SearchQueryStat]:
        buckets = self._date_buckets_between(start_date, end_date)
        agg: dict[str, dict] = defaultdict(lambda: {"count": 0, "total_results": 0})
        for bucket in buckets:
            rows = self._session.execute(
                f"SELECT query, result_count FROM {self._keyspace}.search_events "  # noqa: S608
                f"WHERE date_bucket = %s AND timestamp >= %s AND timestamp <= %s "
                f"ALLOW FILTERING",
                (bucket, start_date, end_date),
            )
            for row in rows:
                agg[row.query]["count"] += 1
                agg[row.query]["total_results"] += row.result_count

        stats = sorted(agg.items(), key=lambda x: x[1]["count"], reverse=True)
        return [
            SearchQueryStat(
                query=q,
                searchCount=data["count"],
                avgResultCount=data["total_results"] / data["count"],
            )
            for q, data in stats[:limit]
        ]

    async def get_summary(
        self,
        start_date: datetime,
        end_date: datetime,
    ) -> AnalyticsSummary:
        buckets = self._date_buckets_between(start_date, end_date)
        total_pv = 0
        total_pc = 0
        total_sr = 0
        sessions: set = set()

        for bucket in buckets:
            pv_rows = self._session.execute(
                f"SELECT session_id FROM {self._keyspace}.page_views "  # noqa: S608
                f"WHERE date_bucket = %s AND timestamp >= %s AND timestamp <= %s ALLOW FILTERING",
                (bucket, start_date, end_date),
            )
            for row in pv_rows:
                total_pv += 1
                sessions.add(row.session_id)

            pc_rows = self._session.execute(
                f"SELECT session_id FROM {self._keyspace}.product_clicks "  # noqa: S608
                f"WHERE date_bucket = %s AND timestamp >= %s AND timestamp <= %s ALLOW FILTERING",
                (bucket, start_date, end_date),
            )
            for row in pc_rows:
                total_pc += 1
                sessions.add(row.session_id)

            sr_rows = self._session.execute(
                f"SELECT session_id FROM {self._keyspace}.search_events "  # noqa: S608
                f"WHERE date_bucket = %s AND timestamp >= %s AND timestamp <= %s ALLOW FILTERING",
                (bucket, start_date, end_date),
            )
            for row in sr_rows:
                total_sr += 1
                sessions.add(row.session_id)

        return AnalyticsSummary(
            totalPageViews=total_pv,
            totalProductClicks=total_pc,
            totalSearches=total_sr,
            uniqueSessions=len(sessions),
            startDate=start_date,
            endDate=end_date,
        )
