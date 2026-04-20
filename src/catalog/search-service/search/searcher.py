import logging
import time
from typing import Any

from elasticsearch import AsyncElasticsearch, NotFoundError

from search.config import settings
from search.models import ProductDocument, SearchRequest, SearchResult

logger = logging.getLogger(__name__)


def _build_query(req: SearchRequest) -> dict[str, Any]:
    """Build the Elasticsearch query body from a SearchRequest."""
    must: list[dict[str, Any]] = []
    filters: list[dict[str, Any]] = []

    # Full-text search across name_text, description, sku, and tags
    if req.q:
        must.append(
            {
                "multi_match": {
                    "query": req.q,
                    "fields": ["name_text^3", "description", "sku^2", "tags^1"],
                    "type": "best_fields",
                    "fuzziness": "AUTO",
                    "operator": "or",
                }
            }
        )
    else:
        must.append({"match_all": {}})

    # Filters
    if req.status:
        filters.append({"term": {"status": req.status}})

    if req.category_id:
        filters.append({"term": {"category_id": req.category_id}})

    if req.brand_id:
        filters.append({"term": {"brand_id": req.brand_id}})

    if req.tags:
        filters.append({"terms": {"tags": req.tags}})

    price_range: dict[str, float] = {}
    if req.min_price > 0:
        price_range["gte"] = req.min_price
    if req.max_price > 0:
        price_range["lte"] = req.max_price
    if price_range:
        filters.append({"range": {"price": price_range}})

    query: dict[str, Any] = {
        "bool": {
            "must": must,
            "filter": filters,
        }
    }
    return query


def _build_aggs() -> dict[str, Any]:
    """Build standard aggregations for faceted filtering."""
    return {
        "by_category": {
            "terms": {"field": "category_id", "size": 50}
        },
        "by_brand": {
            "terms": {"field": "brand_id", "size": 50}
        },
        "price_range": {
            "range": {
                "field": "price",
                "ranges": [
                    {"key": "under_25", "to": 25},
                    {"key": "25_to_50", "from": 25, "to": 50},
                    {"key": "50_to_100", "from": 50, "to": 100},
                    {"key": "100_to_250", "from": 100, "to": 250},
                    {"key": "over_250", "from": 250},
                ],
            }
        },
        "price_stats": {
            "stats": {"field": "price"}
        },
    }


def _resolve_sort(sort: str) -> list[Any]:
    """Map sort string to Elasticsearch sort spec."""
    sort_map: dict[str, Any] = {
        "_score": [{"_score": {"order": "desc"}}],
        "price_asc": [{"price": {"order": "asc"}}],
        "price_desc": [{"price": {"order": "desc"}}],
        "name_asc": [{"name.keyword": {"order": "asc"}}],
        "newest": [{"created_at": {"order": "desc"}}],
    }
    return sort_map.get(sort, [{"_score": {"order": "desc"}}])


def _parse_hit(hit: dict[str, Any]) -> ProductDocument:
    """Convert a raw ES hit into a ProductDocument."""
    src = hit["_source"]
    # name field is stored as a completion object; recover plain text
    raw_name = src.get("name_text") or src.get("name")
    if isinstance(raw_name, dict):
        raw_name = raw_name.get("input", [""])[0]
    src = dict(src)
    src["name"] = raw_name
    return ProductDocument(**src)


def _parse_aggs(raw: dict[str, Any]) -> dict[str, Any]:
    """Simplify raw ES aggregation response."""
    result: dict[str, Any] = {}

    if "by_category" in raw:
        result["by_category"] = [
            {"key": b["key"], "count": b["doc_count"]}
            for b in raw["by_category"].get("buckets", [])
        ]

    if "by_brand" in raw:
        result["by_brand"] = [
            {"key": b["key"], "count": b["doc_count"]}
            for b in raw["by_brand"].get("buckets", [])
        ]

    if "price_range" in raw:
        result["price_range"] = [
            {
                "key": b.get("key", ""),
                "from": b.get("from"),
                "to": b.get("to"),
                "count": b["doc_count"],
            }
            for b in raw["price_range"].get("buckets", [])
        ]

    if "price_stats" in raw:
        result["price_stats"] = raw["price_stats"]

    return result


class ProductSearcher:
    def __init__(self, client: AsyncElasticsearch | None = None) -> None:
        self._client = client or AsyncElasticsearch(
            hosts=[settings.elasticsearch_url]
        )
        self._index = settings.index_name

    @property
    def client(self) -> AsyncElasticsearch:
        return self._client

    async def search(self, req: SearchRequest) -> SearchResult:
        t0 = time.monotonic()

        body: dict[str, Any] = {
            "query": _build_query(req),
            "aggs": _build_aggs(),
            "from": req.from_,
            "size": req.size,
            "sort": _resolve_sort(req.sort),
            "track_total_hits": True,
        }

        response = await self._client.search(index=self._index, body=body)

        took_ms = int((time.monotonic() - t0) * 1000)
        total_val = response["hits"]["total"]
        total = total_val["value"] if isinstance(total_val, dict) else int(total_val)

        hits = [_parse_hit(h) for h in response["hits"]["hits"]]
        aggs = _parse_aggs(response.get("aggregations", {}))

        return SearchResult(total=total, hits=hits, aggregations=aggs, took_ms=took_ms)

    async def suggest(self, prefix: str, size: int = 5) -> list[str]:
        """Return name suggestions for the given prefix using a prefix query."""
        if not prefix:
            return []

        body: dict[str, Any] = {
            "suggest": {
                "product_suggest": {
                    "prefix": prefix,
                    "completion": {
                        "field": "name",
                        "size": size,
                        "skip_duplicates": True,
                        "fuzzy": {"fuzziness": 1},
                    },
                }
            }
        }

        try:
            response = await self._client.search(index=self._index, body=body)
        except Exception:
            # Fall back to a plain prefix match query
            return await self._prefix_fallback(prefix, size)

        suggestions: list[str] = []
        for entry in response.get("suggest", {}).get("product_suggest", []):
            for option in entry.get("options", []):
                text = option.get("_source", {}).get("name_text") or option.get("text", "")
                if text and text not in suggestions:
                    suggestions.append(text)

        if not suggestions:
            return await self._prefix_fallback(prefix, size)

        return suggestions[:size]

    async def _prefix_fallback(self, prefix: str, size: int) -> list[str]:
        """Fallback prefix search using a match_phrase_prefix query on name_text."""
        body: dict[str, Any] = {
            "query": {
                "match_phrase_prefix": {
                    "name_text": {"query": prefix, "max_expansions": 20}
                }
            },
            "size": size,
            "_source": ["name_text"],
        }
        response = await self._client.search(index=self._index, body=body)
        results: list[str] = []
        for hit in response["hits"]["hits"]:
            name = hit["_source"].get("name_text", "")
            if name and name not in results:
                results.append(name)
        return results[:size]

    async def get_by_id(self, product_id: str) -> ProductDocument | None:
        try:
            response = await self._client.get(index=self._index, id=product_id)
            return _parse_hit(response)
        except NotFoundError:
            return None

    async def close(self) -> None:
        await self._client.close()
