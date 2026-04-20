"""Integration-style tests for the FastAPI router — searcher/indexer are mocked."""
import sys
import os
from datetime import datetime, timezone
from unittest.mock import AsyncMock, MagicMock

import pytest
from fastapi import FastAPI
from fastapi.testclient import TestClient

# Ensure the project root is on sys.path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

from handler.api import router
from search.models import ProductDocument, SearchResult


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


def _sample_product(product_id: str = "prod-1") -> ProductDocument:
    return ProductDocument(
        id=product_id,
        sku="SKU-001",
        name="Blue Widget",
        description="A fine blue widget",
        category_id="cat-electronics",
        brand_id="brand-acme",
        brand_name="Acme",
        price=49.99,
        currency="USD",
        status="active",
        tags=["blue", "widget"],
        attributes={"color": "blue"},
        image_urls=["https://example.com/img.jpg"],
        created_at=datetime(2024, 1, 15, tzinfo=timezone.utc),
    )


def _sample_search_result(product_id: str = "prod-1") -> SearchResult:
    return SearchResult(
        total=1,
        hits=[_sample_product(product_id)],
        aggregations={
            "by_category": [{"key": "cat-electronics", "count": 1}],
            "by_brand": [{"key": "brand-acme", "count": 1}],
        },
        took_ms=5,
    )


def _build_app(mock_searcher: AsyncMock, mock_indexer: AsyncMock) -> FastAPI:
    """Build a minimal FastAPI app with mocked state."""
    app = FastAPI()
    app.include_router(router)
    app.state.searcher = mock_searcher
    app.state.indexer = mock_indexer
    return app


# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------


@pytest.fixture()
def mock_searcher() -> AsyncMock:
    searcher = AsyncMock()
    searcher.search = AsyncMock(return_value=_sample_search_result())
    searcher.suggest = AsyncMock(return_value=["Blue Widget", "Blue Gadget"])
    searcher.get_by_id = AsyncMock(return_value=_sample_product())
    return searcher


@pytest.fixture()
def mock_indexer() -> AsyncMock:
    indexer = AsyncMock()
    indexer.index_product = AsyncMock(return_value=None)
    indexer.delete_product = AsyncMock(return_value=None)
    indexer.bulk_index = AsyncMock(return_value=(2, []))
    return indexer


@pytest.fixture()
def client(mock_searcher: AsyncMock, mock_indexer: AsyncMock) -> TestClient:
    app = _build_app(mock_searcher, mock_indexer)
    return TestClient(app)


# ---------------------------------------------------------------------------
# GET /healthz
# ---------------------------------------------------------------------------


class TestHealthz:
    def test_returns_200(self, client: TestClient) -> None:
        response = client.get("/healthz")
        assert response.status_code == 200

    def test_returns_ok_status(self, client: TestClient) -> None:
        response = client.get("/healthz")
        assert response.json() == {"status": "ok"}


# ---------------------------------------------------------------------------
# POST /search/products
# ---------------------------------------------------------------------------


class TestSearchProducts:
    def test_returns_200(self, client: TestClient) -> None:
        response = client.post("/search/products", json={"q": "widget"})
        assert response.status_code == 200

    def test_response_shape_matches_search_result(self, client: TestClient) -> None:
        response = client.post("/search/products", json={"q": "widget"})
        data = response.json()
        assert "total" in data
        assert "hits" in data
        assert "aggregations" in data
        assert "took_ms" in data

    def test_hits_contain_product_fields(self, client: TestClient) -> None:
        response = client.post("/search/products", json={"q": "widget"})
        hit = response.json()["hits"][0]
        assert hit["id"] == "prod-1"
        assert hit["name"] == "Blue Widget"
        assert hit["price"] == 49.99

    def test_empty_query_accepted(self, client: TestClient) -> None:
        response = client.post("/search/products", json={})
        assert response.status_code == 200

    def test_searcher_called_once(
        self, client: TestClient, mock_searcher: AsyncMock
    ) -> None:
        client.post("/search/products", json={"q": "test"})
        mock_searcher.search.assert_called_once()

    def test_searcher_error_returns_500(
        self, mock_searcher: AsyncMock, mock_indexer: AsyncMock
    ) -> None:
        mock_searcher.search = AsyncMock(side_effect=RuntimeError("ES down"))
        app = _build_app(mock_searcher, mock_indexer)
        with TestClient(app) as c:
            response = c.post("/search/products", json={"q": "crash"})
        assert response.status_code == 500

    def test_with_filters(self, client: TestClient) -> None:
        payload = {
            "q": "widget",
            "category_id": "cat-electronics",
            "brand_id": "brand-acme",
            "min_price": 10.0,
            "max_price": 100.0,
            "status": "active",
            "from": 0,
            "size": 10,
        }
        response = client.post("/search/products", json=payload)
        assert response.status_code == 200


# ---------------------------------------------------------------------------
# GET /search/suggest
# ---------------------------------------------------------------------------


class TestSuggest:
    def test_returns_200(self, client: TestClient) -> None:
        response = client.get("/search/suggest", params={"prefix": "blue"})
        assert response.status_code == 200

    def test_returns_list_of_strings(self, client: TestClient) -> None:
        response = client.get("/search/suggest", params={"prefix": "blue"})
        data = response.json()
        assert isinstance(data, list)
        assert all(isinstance(s, str) for s in data)

    def test_suggestions_content(self, client: TestClient) -> None:
        response = client.get("/search/suggest", params={"prefix": "blue"})
        assert "Blue Widget" in response.json()

    def test_missing_prefix_returns_422(self, client: TestClient) -> None:
        response = client.get("/search/suggest")
        assert response.status_code == 422

    def test_custom_size_param(
        self, client: TestClient, mock_searcher: AsyncMock
    ) -> None:
        client.get("/search/suggest", params={"prefix": "wid", "size": "3"})
        mock_searcher.suggest.assert_called_once_with(prefix="wid", size=3)

    def test_suggest_error_returns_500(
        self, mock_searcher: AsyncMock, mock_indexer: AsyncMock
    ) -> None:
        mock_searcher.suggest = AsyncMock(side_effect=RuntimeError("ES down"))
        app = _build_app(mock_searcher, mock_indexer)
        with TestClient(app) as c:
            response = c.get("/search/suggest", params={"prefix": "crash"})
        assert response.status_code == 500


# ---------------------------------------------------------------------------
# POST /search/index
# ---------------------------------------------------------------------------


class TestIndexProduct:
    def _product_payload(self, product_id: str = "prod-new") -> dict:
        return {
            "id": product_id,
            "sku": "NEW-001",
            "name": "New Product",
            "description": "Brand new",
            "category_id": "cat-new",
            "brand_id": "brand-new",
            "brand_name": "NewBrand",
            "price": 19.99,
            "currency": "USD",
            "status": "active",
            "tags": [],
            "attributes": {},
            "image_urls": [],
            "created_at": "2024-06-01T00:00:00Z",
        }

    def test_returns_201(self, client: TestClient) -> None:
        response = client.post("/search/index", json=self._product_payload())
        assert response.status_code == 201

    def test_response_contains_id(self, client: TestClient) -> None:
        response = client.post("/search/index", json=self._product_payload("p-123"))
        data = response.json()
        assert data["id"] == "p-123"
        assert data["result"] == "indexed"

    def test_indexer_called_once(
        self, client: TestClient, mock_indexer: AsyncMock
    ) -> None:
        client.post("/search/index", json=self._product_payload())
        mock_indexer.index_product.assert_called_once()

    def test_indexer_error_returns_500(
        self, mock_searcher: AsyncMock, mock_indexer: AsyncMock
    ) -> None:
        mock_indexer.index_product = AsyncMock(side_effect=RuntimeError("ES down"))
        app = _build_app(mock_searcher, mock_indexer)
        with TestClient(app) as c:
            response = c.post("/search/index", json=self._product_payload())
        assert response.status_code == 500


# ---------------------------------------------------------------------------
# DELETE /search/index/{product_id}
# ---------------------------------------------------------------------------


class TestDeleteFromIndex:
    def test_returns_204(self, client: TestClient) -> None:
        response = client.delete("/search/index/prod-1")
        assert response.status_code == 204

    def test_indexer_delete_called(
        self, client: TestClient, mock_indexer: AsyncMock
    ) -> None:
        client.delete("/search/index/prod-xyz")
        mock_indexer.delete_product.assert_called_once_with("prod-xyz")

    def test_delete_error_returns_500(
        self, mock_searcher: AsyncMock, mock_indexer: AsyncMock
    ) -> None:
        mock_indexer.delete_product = AsyncMock(side_effect=RuntimeError("ES down"))
        app = _build_app(mock_searcher, mock_indexer)
        with TestClient(app) as c:
            response = c.delete("/search/index/crash")
        assert response.status_code == 500


# ---------------------------------------------------------------------------
# POST /search/bulk-index
# ---------------------------------------------------------------------------


class TestBulkIndex:
    def _two_products(self) -> list[dict]:
        base = {
            "sku": "SKU-X",
            "name": "Product",
            "description": "Desc",
            "category_id": "c1",
            "brand_id": "b1",
            "brand_name": "Brand",
            "price": 9.99,
            "status": "active",
            "created_at": "2024-01-01T00:00:00Z",
        }
        return [
            {**base, "id": "p1"},
            {**base, "id": "p2"},
        ]

    def test_returns_200(self, client: TestClient) -> None:
        response = client.post("/search/bulk-index", json=self._two_products())
        assert response.status_code == 200

    def test_returns_indexed_count(self, client: TestClient) -> None:
        response = client.post("/search/bulk-index", json=self._two_products())
        data = response.json()
        assert data["indexed"] == 2
        assert data["errors"] == 0

    def test_empty_list_returns_zero(self, client: TestClient) -> None:
        response = client.post("/search/bulk-index", json=[])
        assert response.status_code == 200
        data = response.json()
        assert data["indexed"] == 0

    def test_bulk_index_error_returns_500(
        self, mock_searcher: AsyncMock, mock_indexer: AsyncMock
    ) -> None:
        mock_indexer.bulk_index = AsyncMock(side_effect=RuntimeError("ES down"))
        app = _build_app(mock_searcher, mock_indexer)
        with TestClient(app) as c:
            response = c.post("/search/bulk-index", json=self._two_products())
        assert response.status_code == 500
