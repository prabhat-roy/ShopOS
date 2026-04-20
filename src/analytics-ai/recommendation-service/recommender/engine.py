from __future__ import annotations

import logging
from typing import Optional

import numpy as np

from .models import ProductRec, RecommendRequest, RecommendResponse
from .store import InteractionStore, interaction_store

logger = logging.getLogger(__name__)


def _jaccard(set_a: set[str], set_b: set[str]) -> float:
    """Jaccard similarity between two sets. Returns 0.0 when both are empty."""
    if not set_a and not set_b:
        return 0.0
    intersection = len(set_a & set_b)
    union = len(set_a | set_b)
    return intersection / union if union > 0 else 0.0


class RecommendationEngine:
    """
    Pure in-memory recommendation engine backed by an InteractionStore.

    Strategies:
      - recommend_for_user   : user-based collaborative filtering (Jaccard)
      - recommend_similar    : item-based collaborative filtering (co-purchase)
      - recommend_popular    : popularity fallback
      - recommend            : hybrid dispatcher
    """

    def __init__(self, store: InteractionStore) -> None:
        self._store = store

    # ------------------------------------------------------------------
    # Public recommendation methods
    # ------------------------------------------------------------------

    def recommend_for_user(
        self,
        userId: str,
        limit: int = 10,
    ) -> list[ProductRec]:
        """
        User-based collaborative filtering.

        1. Collect *purchase* (and fallback *all*) product sets for all users.
        2. Compute Jaccard similarity between *userId*'s set and every other user.
        3. Pick top-N similar users, aggregate their products.
        4. Exclude products already seen by *userId*.
        """
        user_history = self._store.get_user_history(userId)
        all_users = self._store.get_all_users()

        if not all_users:
            return self.recommend_popular(limit=limit, excludeProductIds=user_history)

        # Build purchase sets per user (weighted by interaction type)
        user_purchase_sets: dict[str, set[str]] = {}
        for uid in all_users:
            interactions = self._store.get_user_interactions(uid)
            # Use all interaction types so sparse data still yields similarity
            user_purchase_sets[uid] = {i.productId for i in interactions}

        target_set = user_purchase_sets.get(userId, set())

        # Compute similarity against all other users
        similarities: list[tuple[str, float]] = []
        for uid in all_users:
            if uid == userId:
                continue
            sim = _jaccard(target_set, user_purchase_sets[uid])
            if sim > 0:
                similarities.append((uid, sim))

        similarities.sort(key=lambda x: x[1], reverse=True)
        # Consider top-20 similar users
        top_peers = similarities[:20]

        if not top_peers:
            return self.recommend_popular(limit=limit, excludeProductIds=user_history)

        # Weighted vote for each product from peer users
        product_scores: dict[str, float] = {}
        for uid, sim in top_peers:
            for interaction in self._store.get_user_interactions(uid):
                pid = interaction.productId
                if pid in user_history:
                    continue  # already seen by this user
                product_scores[pid] = product_scores.get(pid, 0.0) + sim

        if not product_scores:
            return self.recommend_popular(limit=limit, excludeProductIds=user_history)

        sorted_products = sorted(product_scores.items(), key=lambda kv: kv[1], reverse=True)

        results: list[ProductRec] = []
        for pid, score in sorted_products[:limit]:
            results.append(
                ProductRec(
                    productId=pid,
                    score=round(score, 4),
                    reason="Recommended based on similar users' preferences",
                )
            )
        return results

    def recommend_similar(
        self,
        productId: str,
        limit: int = 10,
    ) -> list[ProductRec]:
        """
        Item-based collaborative filtering using co-purchase signals.
        """
        co_products = self._store.get_co_purchased(productId, limit=limit * 2)

        if not co_products:
            return self.recommend_popular(
                limit=limit, excludeProductIds={productId}
            )

        results: list[ProductRec] = []
        for i, pid in enumerate(co_products[:limit]):
            # Score decays with rank; rank 0 → 1.0, rank N → approaches 0
            score = round(1.0 / (i + 1), 4)
            results.append(
                ProductRec(
                    productId=pid,
                    score=score,
                    reason=f"Frequently bought together with {productId}",
                )
            )
        return results

    def recommend_popular(
        self,
        limit: int = 10,
        excludeProductIds: Optional[set[str]] = None,
    ) -> list[ProductRec]:
        """
        Popularity-based fallback: products with the highest cumulative
        interaction weight.
        """
        exclude = excludeProductIds or set()
        popular = self._store.get_popular_products(limit=limit + len(exclude) + 10)

        results: list[ProductRec] = []
        for i, pid in enumerate(popular):
            if pid in exclude:
                continue
            score = round(1.0 / (i + 1), 4)
            results.append(
                ProductRec(
                    productId=pid,
                    score=score,
                    reason="Trending product",
                )
            )
            if len(results) >= limit:
                break

        return results

    def recommend(self, req: RecommendRequest) -> RecommendResponse:
        """
        Dispatch to the appropriate strategy.

        hybrid:
          - If userId provided → user-based collaborative filtering,
            supplemented with popular items if results are sparse.
          - If productId provided (no userId) → item-based co-purchase.
          - If neither → popular.

        user-based  → recommend_for_user only
        item-based  → recommend_similar only
        popular     → recommend_popular only
        """
        strategy = req.strategy
        limit = req.limit

        if strategy == "popular":
            recs = self.recommend_popular(limit=limit)
            used_strategy = "popular"

        elif strategy == "user-based":
            if not req.userId:
                recs = self.recommend_popular(limit=limit)
                used_strategy = "popular"
            else:
                recs = self.recommend_for_user(req.userId, limit=limit)
                used_strategy = "user-based"

        elif strategy == "item-based":
            if not req.productId:
                recs = self.recommend_popular(limit=limit)
                used_strategy = "popular"
            else:
                recs = self.recommend_similar(req.productId, limit=limit)
                used_strategy = "item-based"

        else:  # hybrid
            if req.userId:
                recs = self.recommend_for_user(req.userId, limit=limit)
                used_strategy = "hybrid:user-based"
                # Supplement with popular if results are thin
                if len(recs) < limit:
                    seen = {r.productId for r in recs}
                    user_history = self._store.get_user_history(req.userId)
                    popular_recs = self.recommend_popular(
                        limit=limit - len(recs),
                        excludeProductIds=seen | user_history,
                    )
                    recs.extend(popular_recs)
            elif req.productId:
                recs = self.recommend_similar(req.productId, limit=limit)
                used_strategy = "hybrid:item-based"
            else:
                recs = self.recommend_popular(limit=limit)
                used_strategy = "hybrid:popular"

        return RecommendResponse(
            recommendations=recs[:limit],
            strategy=used_strategy,
            userId=req.userId,
            productId=req.productId,
        )


# Module-level singleton
recommendation_engine = RecommendationEngine(store=interaction_store)
