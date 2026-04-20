from __future__ import annotations

from datetime import datetime, timedelta
from unittest.mock import AsyncMock, MagicMock

import pytest
from fastapi import FastAPI
from fastapi.testclient import TestClient

from analytics.api import router
from analytics.models import (
    AnalyticsSummary,
    PageViewRecord,
    SearchQueryStat,
    TopProduct,
)
from analytics.store import InMemoryAnalyticsStore


def _build_test_app(store) -> FastAPI:
    app = FastAPI()
    app.include_router(router)
    app.state.store = store
    return app


@pytest.fixture
def store() -> InMemoryAnalyticsStore:
    return InMemoryAnalyticsStore()


@pytest.fixture
def client(store: InMemoryAnalyticsStore) -> TestClient:
    app = _build_test_app(store)
    return TestClient(app)


@pytest.fixture
def mock_store() -> MagicMock:
    s = MagicMock(spec=InMemoryAnalyticsStore)
    s.get_page_views = AsyncMock(return_value=[])
    s.get_top_products = AsyncMock(return_value=[])
    s.get_search_queries = AsyncMock(return_value=[])
    s.get_summary = AsyncMock(
        return_value=AnalyticsSummary(
            totalPageViews=0,
            totalProductClicks=0,
            totalSearches=0,
            uniqueSessions=0,
            startDate=datetime(2024, 1, 1),
            endDate=datetime(2024, 1, 2),
        )
    )
    return s


@pytest.fixture
def mock_client(mock_store: MagicMock) -> TestClient:
    app = _build_test_app(mock_store)
    return TestClient(app)


def test_healthz(client: TestClient) -> None:
    response = client.get("/healthz")
    assert response.status_code == 200
    body = response.json()
    assert body["status"] == "ok"
    assert body["service"] == "analytics-service"


def test_page_views_returns_200(mock_client: TestClient) -> None:
    response = mock_client.get("/analytics/page-views?start=2024-01-01&end=2024-01-31")
    assert response.status_code == 200
    assert isinstance(response.json(), list)


def test_page_views_with_records(store: InMemoryAnalyticsStore) -> None:
    import asyncio

    from analytics.models import PageViewEvent

    asyncio.get_event_loop().run_until_complete(
        store.save_page_view(
            PageViewEvent(
                sessionId="sess-test",
                userId="u1",
                pageUrl="/products",
                timestamp=datetime(2024, 6, 10, 10, 0, 0),
            )
        )
    )

    app = _build_test_app(store)
    client = TestClient(app)
    response = client.get("/analytics/page-views?start=2024-06-01&end=2024-06-30")
    assert response.status_code == 200
    data = response.json()
    assert len(data) == 1
    assert data[0]["pageUrl"] == "/products"


def test_top_products_returns_200(mock_client: TestClient) -> None:
    response = mock_client.get("/analytics/top-products?start=2024-01-01&end=2024-01-31")
    assert response.status_code == 200
    assert isinstance(response.json(), list)


def test_searches_returns_200(mock_client: TestClient) -> None:
    response = mock_client.get("/analytics/searches?start=2024-01-01&end=2024-01-31")
    assert response.status_code == 200
    assert isinstance(response.json(), list)


def test_summary_returns_correct_shape(mock_client: TestClient) -> None:
    response = mock_client.get("/analytics/summary?start=2024-01-01&end=2024-01-02")
    assert response.status_code == 200
    body = response.json()
    assert "totalPageViews" in body
    assert "totalProductClicks" in body
    assert "totalSearches" in body
    assert "uniqueSessions" in body


def test_invalid_date_format_returns_422(client: TestClient) -> None:
    response = client.get("/analytics/page-views?start=not-a-date&end=2024-01-31")
    assert response.status_code == 422


def test_end_before_start_returns_422(client: TestClient) -> None:
    response = client.get("/analytics/page-views?start=2024-06-30&end=2024-06-01")
    assert response.status_code == 422
