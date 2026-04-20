from __future__ import annotations

import threading
from collections import defaultdict
from typing import Optional

from .models import UserInteraction

# Interaction type weights for scoring
INTERACTION_WEIGHTS: dict[str, float] = {
    "purchase": 5.0,
    "cart": 3.0,
    "wishlist": 2.0,
    "view": 1.0,
}


class InteractionStore:
    """
    Thread-safe in-memory store for user–product interactions.

    Internal structures:
      _user_interactions : userId  → list[UserInteraction]
      _product_interactions : productId → list[UserInteraction]
      _product_scores : productId → cumulative weighted score (for popularity)
    """

    def __init__(self) -> None:
        self._lock = threading.RLock()
        self._user_interactions: dict[str, list[UserInteraction]] = defaultdict(list)
        self._product_interactions: dict[str, list[UserInteraction]] = defaultdict(list)
        self._product_scores: dict[str, float] = defaultdict(float)

    # ------------------------------------------------------------------
    # Write
    # ------------------------------------------------------------------

    def add_interaction(self, interaction: UserInteraction) -> None:
        weight = INTERACTION_WEIGHTS.get(interaction.interactionType, 1.0)
        weighted_score = interaction.score * weight

        with self._lock:
            self._user_interactions[interaction.userId].append(interaction)
            self._product_interactions[interaction.productId].append(interaction)
            self._product_scores[interaction.productId] += weighted_score

    # ------------------------------------------------------------------
    # Read — user
    # ------------------------------------------------------------------

    def get_user_interactions(self, userId: str) -> list[UserInteraction]:
        with self._lock:
            return list(self._user_interactions.get(userId, []))

    def get_user_history(self, userId: str) -> set[str]:
        """Return the set of productIds this user has already interacted with."""
        with self._lock:
            return {i.productId for i in self._user_interactions.get(userId, [])}

    def get_all_users(self) -> list[str]:
        with self._lock:
            return list(self._user_interactions.keys())

    # ------------------------------------------------------------------
    # Read — product
    # ------------------------------------------------------------------

    def get_product_interactions(self, productId: str) -> list[UserInteraction]:
        with self._lock:
            return list(self._product_interactions.get(productId, []))

    def get_popular_products(self, limit: int = 20) -> list[str]:
        """Return productIds sorted by cumulative weighted interaction score."""
        with self._lock:
            sorted_products = sorted(
                self._product_scores.items(),
                key=lambda kv: kv[1],
                reverse=True,
            )
            return [pid for pid, _ in sorted_products[:limit]]

    def get_co_purchased(self, productId: str, limit: int = 20) -> list[str]:
        """
        Return products frequently bought together with *productId*.

        Strategy: find all users who purchased *productId*, then count how
        often each other product appears in those users' purchase histories.
        """
        with self._lock:
            # Users who interacted with the target product
            buyer_ids: set[str] = {
                i.userId
                for i in self._product_interactions.get(productId, [])
                if i.interactionType == "purchase"
            }

            if not buyer_ids:
                # Fall back to any interaction type when purchases are scarce
                buyer_ids = {
                    i.userId
                    for i in self._product_interactions.get(productId, [])
                }

            co_counts: dict[str, float] = defaultdict(float)
            for uid in buyer_ids:
                for interaction in self._user_interactions.get(uid, []):
                    if interaction.productId == productId:
                        continue
                    weight = INTERACTION_WEIGHTS.get(interaction.interactionType, 1.0)
                    co_counts[interaction.productId] += weight

        sorted_co = sorted(co_counts.items(), key=lambda kv: kv[1], reverse=True)
        return [pid for pid, _ in sorted_co[:limit]]

    # ------------------------------------------------------------------
    # Utility
    # ------------------------------------------------------------------

    def total_interactions(self) -> int:
        with self._lock:
            return sum(len(v) for v in self._user_interactions.values())

    def clear(self) -> None:
        """Remove all data — useful for tests."""
        with self._lock:
            self._user_interactions.clear()
            self._product_interactions.clear()
            self._product_scores.clear()


# Module-level singleton shared across the application
interaction_store = InteractionStore()
