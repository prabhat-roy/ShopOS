import logging
from typing import Any

from elasticsearch import AsyncElasticsearch
from elasticsearch.helpers import async_bulk

from search.config import settings
from search.models import ProductDocument

logger = logging.getLogger(__name__)

INDEX_MAPPING: dict[str, Any] = {
    "mappings": {
        "properties": {
            "id": {"type": "keyword"},
            "sku": {"type": "keyword"},
            "name": {
                "type": "text",
                "analyzer": "standard",
                "fields": {
                    "keyword": {"type": "keyword"},
                    "suggest": {
                        "type": "completion",
                    },
                },
            },
            "description": {"type": "text", "analyzer": "standard"},
            "category_id": {"type": "keyword"},
            "brand_id": {"type": "keyword"},
            "brand_name": {
                "type": "text",
                "analyzer": "standard",
                "fields": {"keyword": {"type": "keyword"}},
            },
            "price": {"type": "float"},
            "currency": {"type": "keyword"},
            "status": {"type": "keyword"},
            "tags": {"type": "keyword"},
            "attributes": {"type": "object", "dynamic": True},
            "image_urls": {"type": "keyword", "index": False},
            "created_at": {"type": "date"},
        }
    },
    "settings": {
        "number_of_shards": 1,
        "number_of_replicas": 1,
    },
}


class ProductIndexer:
    def __init__(self, client: AsyncElasticsearch | None = None) -> None:
        self._client = client or AsyncElasticsearch(
            hosts=[settings.elasticsearch_url]
        )
        self._index = settings.index_name

    @property
    def client(self) -> AsyncElasticsearch:
        return self._client

    async def create_index_if_not_exists(self) -> None:
        exists = await self._client.indices.exists(index=self._index)
        if not exists:
            await self._client.indices.create(index=self._index, body=INDEX_MAPPING)
            logger.info("Created Elasticsearch index: %s", self._index)
        else:
            logger.info("Elasticsearch index already exists: %s", self._index)

    def _to_doc(self, doc: ProductDocument) -> dict[str, Any]:
        data = doc.model_dump()
        # Store datetime as ISO string; ES will parse the date field
        data["created_at"] = doc.created_at.isoformat()
        # Add completion suggester field for name
        data["name"] = {
            "input": [doc.name] + doc.tags,
            "weight": 1,
        }
        # Keep a plain text copy under name_text for multi_match queries
        data["name_text"] = doc.name
        return data

    async def index_product(self, doc: ProductDocument) -> None:
        body = self._to_doc(doc)
        await self._client.index(
            index=self._index,
            id=doc.id,
            document=body,
            refresh="wait_for",
        )
        logger.debug("Indexed product %s", doc.id)

    async def delete_product(self, product_id: str) -> None:
        await self._client.delete(
            index=self._index,
            id=product_id,
            ignore=[404],
            refresh="wait_for",
        )
        logger.debug("Deleted product %s", product_id)

    async def bulk_index(self, docs: list[ProductDocument]) -> tuple[int, list[Any]]:
        actions = [
            {
                "_op_type": "index",
                "_index": self._index,
                "_id": doc.id,
                **self._to_doc(doc),
            }
            for doc in docs
        ]
        success, errors = await async_bulk(
            self._client, actions, refresh="wait_for", raise_on_error=False
        )
        if errors:
            logger.warning("Bulk index encountered %d errors", len(errors))
        logger.info("Bulk indexed %d products", success)
        return success, errors

    async def close(self) -> None:
        await self._client.close()
