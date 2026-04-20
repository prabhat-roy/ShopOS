from __future__ import annotations

from datetime import datetime, timezone
from typing import Any, Optional

from motor.motor_asyncio import AsyncIOMotorClient, AsyncIOMotorCollection

from .models import UserProfile


class MongoRepository:
    def __init__(self, mongodb_uri: str, db_name: str) -> None:
        self._uri = mongodb_uri
        self._db_name = db_name
        self._client: Optional[AsyncIOMotorClient] = None
        self._collection: Optional[AsyncIOMotorCollection] = None

    async def connect(self) -> None:
        self._client = AsyncIOMotorClient(
            self._uri, serverSelectionTimeoutMS=3000
        )
        db = self._client[self._db_name]
        self._collection = db["user_profiles"]
        await self._collection.create_index("userId", unique=True)

    async def disconnect(self) -> None:
        if self._client:
            self._client.close()

    # ------------------------------------------------------------------
    # Internal helpers
    # ------------------------------------------------------------------

    def _to_profile(self, doc: dict) -> UserProfile:
        doc.pop("_id", None)
        return UserProfile(**doc)

    def _to_doc(self, profile: UserProfile) -> dict:
        return profile.model_dump()

    # ------------------------------------------------------------------
    # Profile CRUD
    # ------------------------------------------------------------------

    async def get_profile(self, user_id: str) -> Optional[UserProfile]:
        doc = await self._collection.find_one({"userId": user_id})
        if doc is None:
            return None
        return self._to_profile(doc)

    async def upsert_profile(self, user_id: str, data: dict) -> UserProfile:
        data["userId"] = user_id
        data["updatedAt"] = datetime.now(timezone.utc)
        await self._collection.update_one(
            {"userId": user_id},
            {"$set": data},
            upsert=True,
        )
        doc = await self._collection.find_one({"userId": user_id})
        return self._to_profile(doc)

    async def update_preferences(
        self,
        user_id: str,
        categories: Optional[list[str]] = None,
        brands: Optional[list[str]] = None,
        price_range: Optional[tuple[float, float]] = None,
    ) -> None:
        updates: dict[str, Any] = {"updatedAt": datetime.now(timezone.utc)}
        if categories is not None:
            updates["preferredCategories"] = categories
        if brands is not None:
            updates["preferredBrands"] = brands
        if price_range is not None:
            updates["priceRangeLow"] = price_range[0]
            updates["priceRangeHigh"] = price_range[1]
        await self._collection.update_one(
            {"userId": user_id},
            {"$set": updates},
            upsert=True,
        )

    async def add_viewed(self, user_id: str, product_id: str) -> None:
        await self._collection.update_one(
            {"userId": user_id},
            {
                "$push": {
                    "recentlyViewedProducts": {
                        "$each": [product_id],
                        "$slice": -50,
                    }
                },
                "$set": {"updatedAt": datetime.now(timezone.utc)},
            },
            upsert=True,
        )

    async def add_purchased(self, user_id: str, product_id: str) -> None:
        await self._collection.update_one(
            {"userId": user_id},
            {
                "$addToSet": {"purchaseHistory": product_id},
                "$set": {"updatedAt": datetime.now(timezone.utc)},
            },
            upsert=True,
        )

    async def delete_profile(self, user_id: str) -> bool:
        result = await self._collection.delete_one({"userId": user_id})
        return result.deleted_count > 0
