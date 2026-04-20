from __future__ import annotations

import asyncio
from datetime import datetime, timedelta

import pytest

from analytics.models import (
    PageViewEvent,
    ProductClickEvent,
    SearchEvent,
)
from analytics.store import InMemoryAnalyticsStore


@pytest.fixture
def store() -> InMemoryAnalyticsStore:
    return InMemoryAnalyticsStore()


@pytest.fixture
def now() -> datetime:
    return datetime(2024, 6, 15, 12, 0, 0)


@pytest.fixture
def window(now: datetime) -> tuple[datetime, datetime]:
    return now - timedelta(hours=1), now + timedelta(hours=1)


@pytest.mark.asyncio
async def test_save_and_retrieve_page_views(
    store: InMemoryAnalyticsStore, now: datetime, window: tuple[datetime, datetime]
) -> None:
    event = PageViewEvent(
        sessionId="sess-1",
        userId="user-1",
        pageUrl="/home",
        referrer="https://google.com",
        userAgent="Mozilla/5.0",
        timestamp=now,
    )
    await store.save_page_view(event)

    start, end = window
    results = await store.get_page_views(start, end, limit=10)
    assert len(results) == 1
    assert results[0].pageUrl == "/home"
    assert results[0].sessionId == "sess-1"


@pytest.mark.asyncio
async def test_page_views_outside_window_excluded(
    store: InMemoryAnalyticsStore, now: datetime
) -> None:
    event = PageViewEvent(
        sessionId="sess-outside",
        userId=None,
        pageUrl="/old-page",
        timestamp=now - timedelta(days=5),
    )
    await store.save_page_view(event)

    start = now - timedelta(hours=1)
    end = now + timedelta(hours=1)
    results = await store.get_page_views(start, end, limit=10)
    assert results == []


@pytest.mark.asyncio
async def test_save_and_retrieve_top_products_sorted_by_count(
    store: InMemoryAnalyticsStore, now: datetime, window: tuple[datetime, datetime]
) -> None:
    for _ in range(3):
        await store.save_product_click(
            ProductClickEvent(sessionId="s1", productId="prod-A", sku="SKU-A", timestamp=now)
        )
    for _ in range(5):
        await store.save_product_click(
            ProductClickEvent(sessionId="s2", productId="prod-B", sku="SKU-B", timestamp=now)
        )
    await store.save_product_click(
        ProductClickEvent(sessionId="s3", productId="prod-C", sku="SKU-C", timestamp=now)
    )

    start, end = window
    top = await store.get_top_products(start, end, limit=10)

    assert len(top) == 3
    assert top[0].productId == "prod-B"
    assert top[0].clickCount == 5
    assert top[1].productId == "prod-A"
    assert top[1].clickCount == 3
    assert top[2].productId == "prod-C"
    assert top[2].clickCount == 1


@pytest.mark.asyncio
async def test_top_products_limit_respected(
    store: InMemoryAnalyticsStore, now: datetime, window: tuple[datetime, datetime]
) -> None:
    for i in range(10):
        await store.save_product_click(
            ProductClickEvent(
                sessionId=f"s{i}", productId=f"prod-{i}", sku=f"SKU-{i}", timestamp=now
            )
        )

    start, end = window
    top = await store.get_top_products(start, end, limit=3)
    assert len(top) == 3


@pytest.mark.asyncio
async def test_save_and_retrieve_search_queries(
    store: InMemoryAnalyticsStore, now: datetime, window: tuple[datetime, datetime]
) -> None:
    for _ in range(4):
        await store.save_search(
            SearchEvent(sessionId="s1", query="blue shoes", resultCount=10, timestamp=now)
        )
    await store.save_search(
        SearchEvent(sessionId="s2", query="red hat", resultCount=5, timestamp=now)
    )

    start, end = window
    stats = await store.get_search_queries(start, end, limit=10)

    assert len(stats) == 2
    assert stats[0].query == "blue shoes"
    assert stats[0].searchCount == 4
    assert stats[0].avgResultCount == 10.0
    assert stats[1].query == "red hat"
    assert stats[1].searchCount == 1


@pytest.mark.asyncio
async def test_summary_counts_correct(
    store: InMemoryAnalyticsStore, now: datetime, window: tuple[datetime, datetime]
) -> None:
    await store.save_page_view(
        PageViewEvent(sessionId="s1", pageUrl="/a", timestamp=now)
    )
    await store.save_page_view(
        PageViewEvent(sessionId="s1", pageUrl="/b", timestamp=now)
    )
    await store.save_product_click(
        ProductClickEvent(sessionId="s2", productId="p1", sku="SKU-1", timestamp=now)
    )
    await store.save_search(
        SearchEvent(sessionId="s3", query="test", resultCount=3, timestamp=now)
    )

    start, end = window
    summary = await store.get_summary(start, end)

    assert summary.totalPageViews == 2
    assert summary.totalProductClicks == 1
    assert summary.totalSearches == 1
    assert summary.uniqueSessions == 3


@pytest.mark.asyncio
async def test_summary_unique_sessions_deduped(
    store: InMemoryAnalyticsStore, now: datetime, window: tuple[datetime, datetime]
) -> None:
    for _ in range(5):
        await store.save_page_view(
            PageViewEvent(sessionId="same-session", pageUrl="/x", timestamp=now)
        )

    start, end = window
    summary = await store.get_summary(start, end)
    assert summary.uniqueSessions == 1


@pytest.mark.asyncio
async def test_summary_empty_range(store: InMemoryAnalyticsStore, now: datetime) -> None:
    future_start = now + timedelta(days=10)
    future_end = now + timedelta(days=11)
    summary = await store.get_summary(future_start, future_end)

    assert summary.totalPageViews == 0
    assert summary.totalProductClicks == 0
    assert summary.totalSearches == 0
    assert summary.uniqueSessions == 0
