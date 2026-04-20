"""Unit tests for ProductSearcher — Elasticsearch client is mocked."""
import sys
import os
from datetime import datetime, timezone
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

# Ensure the project root is on sys.path when running with pytest from any directory
sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

from search.models import SearchRequest
from search.searcher import ProductSearcher, _build_query, _parse_hit, _resolve_sort


# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------


def _sample_hit(product_id: str = "prod-1") -> dict:
    return {
        "_index": "products",
        "_id": product_id,
        "_score": 1.0,
        "_source": {
            "id": product_id,
            "sku": "SKU-001",
            "name": {"input": ["Blue Widget"], "weight": 1},
            "name_text": "Blue Widget",
            "description": "A fine blue widget",
            "category_id": "cat-electronics",
            "brand_id": "brand-acme",
            "brand_name": "Acme",
            "price": 49.99,
            "currency": "USD",
            "status": "active",
            "tags": ["blue", "widget"],
            "attributes": {"color": "blue"},
            "image_urls": [],
            "created_at": "2024-01-15T00:00:00+00:00",
        },
    }


def _mock_search_response(hits: list[dict], total: int = 1) -> dict:
    return {
        "took": 5,
        "hits": {
            "total": {"value": total, "relation": "eq"},
            "hits": hits,
        },
        "aggregations": {
            "by_category": {"buckets": [{"key": "cat-electronics", "doc_count": 1}]},
            "by_brand": {"buckets": [{"key": "brand-acme", "doc_count": 1}]},
            "price_range": {
                "buckets": [
                    {"key": "25_to_50", "from": 25.0, "to": 50.0, "doc_count": 1}
                ]
            },
            "price_stats": {"min": 49.99, "max": 49.99, "avg": 49.99},
        },
    }


# ---------------------------------------------------------------------------
# _build_query
# ---------------------------------------------------------------------------


class TestBuildQuery:
    def test_empty_query_produces_match_all(self):
        req = SearchRequest(q="")
        query = _build_query(req)
        must = query["bool"]["must"]
        assert any("match_all" in clause for clause in must)

    def test_q_produces_multi_match(self):
        req = SearchRequest(q="widget")
        query = _build_query(req)
        must = query["bool"]["must"]
        assert any("multi_match" in clause for clause in must)
        mm = next(c["multi_match"] for c in must if "multi_match" in c)
        assert mm["query"] == "widget"
        assert "fuzziness" in mm

    def test_category_filter_applied(self):
        req = SearchRequest(category_id="cat-123")
        query = _build_query(req)
        filters = query["bool"]["filter"]
        assert any(
            f.get("term", {}).get("category_id") == "cat-123" for f in filters
        )

    def test_brand_filter_applied(self):
        req = SearchRequest(brand_id="brand-xyz")
        query = _build_query(req)
        filters = query["bool"]["filter"]
        assert any(f.get("term", {}).get("brand_id") == "brand-xyz" for f in filters)

    def test_price_range_filter(self):
        req = SearchRequest(min_price=10.0, max_price=50.0)
        query = _build_query(req)
        filters = query["bool"]["filter"]
        range_filters = [f for f in filters if "range" in f]
        assert range_filters
        price_range = range_filters[0]["range"]["price"]
        assert price_range["gte"] == 10.0
        assert price_range["lte"] == 50.0

    def test_price_range_only_min(self):
        req = SearchRequest(min_price=5.0)
        query = _build_query(req)
        filters = query["bool"]["filter"]
        range_filters = [f for f in filters if "range" in f]
        assert range_filters
        price_range = range_filters[0]["range"]["price"]
        assert price_range["gte"] == 5.0
        assert "lte" not in price_range

    def test_tags_filter_applied(self):
        req = SearchRequest(tags=["blue", "sale"])
        query = _build_query(req)
        filters = query["bool"]["filter"]
        assert any("terms" in f and f["terms"].get("tags") == ["blue", "sale"] for f in filters)

    def test_status_filter_always_present(self):
        req = SearchRequest(status="active")
        query = _build_query(req)
        filters = query["bool"]["filter"]
        assert any(f.get("term", {}).get("status") == "active" for f in filters)


# ---------------------------------------------------------------------------
# _resolve_sort
# ---------------------------------------------------------------------------


class TestResolveSort:
    def test_score_sort(self):
        result = _resolve_sort("_score")
        assert result[0]["_score"]["order"] == "desc"

    def test_price_asc(self):
        result = _resolve_sort("price_asc")
        assert result[0]["price"]["order"] == "asc"

    def test_price_desc(self):
        result = _resolve_sort("price_desc")
        assert result[0]["price"]["order"] == "desc"

    def test_newest(self):
        result = _resolve_sort("newest")
        assert result[0]["created_at"]["order"] == "desc"

    def test_unknown_falls_back_to_score(self):
        result = _resolve_sort("unknown_sort")
        assert result[0]["_score"]["order"] == "desc"


# ---------------------------------------------------------------------------
# _parse_hit
# ---------------------------------------------------------------------------


class TestParseHit:
    def test_parse_basic_hit(self):
        hit = _sample_hit("prod-99")
        doc = _parse_hit(hit)
        assert doc.id == "prod-99"
        assert doc.name == "Blue Widget"
        assert doc.price == 49.99

    def test_parse_hit_with_plain_name(self):
        hit = _sample_hit()
        hit["_source"]["name"] = "Plain Name"
        del hit["_source"]["name_text"]
        doc = _parse_hit(hit)
        assert doc.name == "Plain Name"


# ---------------------------------------------------------------------------
# ProductSearcher.search
# ---------------------------------------------------------------------------


@pytest.mark.asyncio
class TestProductSearcherSearch:
    async def test_search_returns_search_result(self):
        mock_client = AsyncMock()
        mock_client.search = AsyncMock(
            return_value=_mock_search_response([_sample_hit()])
        )

        searcher = ProductSearcher(client=mock_client)
        req = SearchRequest(q="widget")
        result = await searcher.search(req)

        assert result.total == 1
        assert len(result.hits) == 1
        assert result.hits[0].id == "prod-1"
        assert result.took_ms >= 0
        assert "by_category" in result.aggregations

    async def test_search_empty_hits(self):
        mock_client = AsyncMock()
        mock_client.search = AsyncMock(
            return_value={
                "took": 2,
                "hits": {"total": {"value": 0, "relation": "eq"}, "hits": []},
                "aggregations": {},
            }
        )
        searcher = ProductSearcher(client=mock_client)
        result = await searcher.search(SearchRequest())
        assert result.total == 0
        assert result.hits == []

    async def test_search_calls_es_with_correct_index(self):
        mock_client = AsyncMock()
        mock_client.search = AsyncMock(
            return_value=_mock_search_response([])
        )
        mock_client.search.return_value["hits"]["total"] = {"value": 0, "relation": "eq"}

        searcher = ProductSearcher(client=mock_client)
        await searcher.search(SearchRequest(q="test"))

        call_kwargs = mock_client.search.call_args
        assert call_kwargs.kwargs.get("index") == "products" or (
            call_kwargs.args and call_kwargs.args[0] == "products"
        )

    async def test_search_pagination(self):
        mock_client = AsyncMock()
        mock_client.search = AsyncMock(
            return_value=_mock_search_response([_sample_hit()])
        )
        searcher = ProductSearcher(client=mock_client)
        req = SearchRequest(from_=10, size=5)
        await searcher.search(req)

        body = mock_client.search.call_args.kwargs["body"]
        assert body["from"] == 10
        assert body["size"] == 5


# ---------------------------------------------------------------------------
# ProductSearcher.suggest
# ---------------------------------------------------------------------------


@pytest.mark.asyncio
class TestProductSearcherSuggest:
    async def test_suggest_returns_list_of_strings(self):
        mock_client = AsyncMock()
        mock_client.search = AsyncMock(
            return_value={
                "suggest": {
                    "product_suggest": [
                        {
                            "text": "bl",
                            "offset": 0,
                            "length": 2,
                            "options": [
                                {
                                    "text": "Blue Widget",
                                    "_source": {"name_text": "Blue Widget"},
                                    "_score": 1.0,
                                }
                            ],
                        }
                    ]
                },
                "hits": {"total": {"value": 0, "relation": "eq"}, "hits": []},
            }
        )
        searcher = ProductSearcher(client=mock_client)
        results = await searcher.suggest(prefix="bl", size=5)
        assert isinstance(results, list)
        assert "Blue Widget" in results

    async def test_suggest_empty_prefix_returns_empty(self):
        mock_client = AsyncMock()
        searcher = ProductSearcher(client=mock_client)
        results = await searcher.suggest(prefix="", size=5)
        assert results == []
        mock_client.search.assert_not_called()

    async def test_suggest_deduplicates_results(self):
        mock_client = AsyncMock()
        mock_client.search = AsyncMock(
            return_value={
                "suggest": {
                    "product_suggest": [
                        {
                            "options": [
                                {"_source": {"name_text": "Blue Widget"}},
                                {"_source": {"name_text": "Blue Widget"}},
                                {"_source": {"name_text": "Blue Gadget"}},
                            ]
                        }
                    ]
                },
                "hits": {"total": {"value": 0, "relation": "eq"}, "hits": []},
            }
        )
        searcher = ProductSearcher(client=mock_client)
        results = await searcher.suggest(prefix="blue", size=5)
        assert results.count("Blue Widget") == 1


# ---------------------------------------------------------------------------
# ProductSearcher.get_by_id
# ---------------------------------------------------------------------------


@pytest.mark.asyncio
class TestProductSearcherGetById:
    async def test_get_by_id_returns_document(self):
        mock_client = AsyncMock()
        mock_client.get = AsyncMock(return_value=_sample_hit("prod-42"))
        searcher = ProductSearcher(client=mock_client)
        doc = await searcher.get_by_id("prod-42")
        assert doc is not None
        assert doc.id == "prod-42"

    async def test_get_by_id_returns_none_for_missing(self):
        from elasticsearch import NotFoundError

        mock_client = AsyncMock()
        mock_client.get = AsyncMock(
            side_effect=NotFoundError(
                message="Not found",
                meta=MagicMock(status=404, headers={}),
                body={"found": False},
            )
        )
        searcher = ProductSearcher(client=mock_client)
        doc = await searcher.get_by_id("does-not-exist")
        assert doc is None
