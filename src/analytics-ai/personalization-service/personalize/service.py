from __future__ import annotations

from typing import Any, Optional

from .engine import PersonalizationEngine
from .models import PersonalizationRequest, PersonalizedResult, UserProfile
from .repository import MongoRepository

_engine = PersonalizationEngine()


class PersonalizationService:
    def __init__(self, repository: MongoRepository) -> None:
        self._repo = repository

    async def personalize(self, req: PersonalizationRequest) -> PersonalizedResult:
        profile = await self._repo.get_profile(req.userId)
        if profile is None:
            profile = _engine.get_default_profile(req.userId)

        ranked_ids, scores = _engine.rank_products(
            profile, req.candidateProductIds, req.limit
        )

        return PersonalizedResult(
            userId=req.userId,
            rankedProductIds=ranked_ids,
            scores=scores,
            contextType=req.contextType,
            profile=profile,
        )

    async def get_profile(self, user_id: str) -> Optional[UserProfile]:
        return await self._repo.get_profile(user_id)

    async def update_profile(self, user_id: str, data: dict[str, Any]) -> UserProfile:
        return await self._repo.upsert_profile(user_id, data)

    async def record_view(self, user_id: str, product_id: str) -> None:
        await self._repo.add_viewed(user_id, product_id)

    async def record_purchase(self, user_id: str, product_id: str) -> None:
        await self._repo.add_purchased(user_id, product_id)

    async def delete_profile(self, user_id: str) -> bool:
        return await self._repo.delete_profile(user_id)
