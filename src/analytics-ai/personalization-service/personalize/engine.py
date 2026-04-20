from __future__ import annotations

from datetime import datetime, timezone
from typing import Optional

from .models import PersonalizedResult, UserProfile


# Scoring weights
WEIGHT_PREFERRED_CATEGORY = 2.0
WEIGHT_PREFERRED_BRAND = 1.5
WEIGHT_EXCLUDED_CATEGORY = -1.0
WEIGHT_RECENTLY_VIEWED = 0.5
WEIGHT_PURCHASED = 1.0


class PersonalizationEngine:
    """
    Rule-based scoring engine that ranks candidate product IDs according to
    how well they align with a user's stored preferences.

    NOTE: In production, product attributes (category, brand) would be
    fetched from the catalog service.  Here we derive them from the
    product ID string via a deterministic convention so the engine can be
    tested without external dependencies:
      * product IDs may encode metadata as "cat:<category>|brand:<brand>|<id>"
    If no metadata is encoded the product receives a neutral base score of 0.
    """

    def _extract_metadata(self, product_id: str) -> tuple[Optional[str], Optional[str]]:
        """Return (category, brand) extracted from an encoded product ID, or (None, None)."""
        category: Optional[str] = None
        brand: Optional[str] = None
        if "|" in product_id:
            for part in product_id.split("|"):
                if part.startswith("cat:"):
                    category = part[4:]
                elif part.startswith("brand:"):
                    brand = part[6:]
        return category, brand

    def _score_product(self, product_id: str, profile: UserProfile) -> float:
        score = 0.0
        category, brand = self._extract_metadata(product_id)

        if category is not None:
            if category in profile.preferredCategories:
                score += WEIGHT_PREFERRED_CATEGORY
            if category in profile.excludedCategories:
                score += WEIGHT_EXCLUDED_CATEGORY

        if brand is not None and brand in profile.preferredBrands:
            score += WEIGHT_PREFERRED_BRAND

        if product_id in profile.recentlyViewedProducts:
            score += WEIGHT_RECENTLY_VIEWED

        if product_id in profile.purchaseHistory:
            score += WEIGHT_PURCHASED

        return score

    def _normalize_scores(self, scores: dict[str, float]) -> dict[str, float]:
        if not scores:
            return {}
        min_s = min(scores.values())
        max_s = max(scores.values())
        span = max_s - min_s
        if span == 0.0:
            return {k: 0.5 for k in scores}
        return {k: (v - min_s) / span for k, v in scores.items()}

    def get_default_profile(self, user_id: str) -> UserProfile:
        return UserProfile(
            userId=user_id,
            preferredCategories=[],
            preferredBrands=[],
            priceRangeLow=0.0,
            priceRangeHigh=10000.0,
            recentlyViewedProducts=[],
            purchaseHistory=[],
            excludedCategories=[],
            updatedAt=datetime.now(timezone.utc),
        )

    def rank_products(
        self,
        profile: UserProfile,
        candidate_ids: list[str],
        limit: int,
    ) -> tuple[list[str], dict[str, float]]:
        """
        Score and rank candidates.  Returns (ranked_ids, normalised_scores).
        """
        if not candidate_ids:
            return [], {}

        raw_scores: dict[str, float] = {
            pid: self._score_product(pid, profile) for pid in candidate_ids
        }

        normalised = self._normalize_scores(raw_scores)
        ranked = sorted(candidate_ids, key=lambda pid: raw_scores[pid], reverse=True)
        ranked = ranked[:limit]

        return ranked, {pid: round(normalised[pid], 6) for pid in ranked}
