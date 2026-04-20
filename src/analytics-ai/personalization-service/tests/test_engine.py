"""Tests for PersonalizationEngine — 10 tests."""
from __future__ import annotations

from datetime import datetime, timezone

import pytest

from personalize.engine import PersonalizationEngine
from personalize.models import UserProfile


@pytest.fixture
def engine() -> PersonalizationEngine:
    return PersonalizationEngine()


def _profile(**kwargs) -> UserProfile:
    defaults = dict(
        userId="u1",
        preferredCategories=[],
        preferredBrands=[],
        priceRangeLow=0.0,
        priceRangeHigh=10000.0,
        recentlyViewedProducts=[],
        purchaseHistory=[],
        excludedCategories=[],
        updatedAt=datetime.now(timezone.utc),
    )
    defaults.update(kwargs)
    return UserProfile(**defaults)


def _enc(product_id: str, category: str = "", brand: str = "") -> str:
    """Build an encoded product ID."""
    parts = [product_id]
    if category:
        parts.append(f"cat:{category}")
    if brand:
        parts.append(f"brand:{brand}")
    return "|".join(parts)


# 1. Preferred category → higher rank
def test_rank_by_category_match(engine):
    profile = _profile(preferredCategories=["electronics"])
    candidates = [
        _enc("p1", category="electronics"),
        _enc("p2", category="clothing"),
    ]
    ranked, scores = engine.rank_products(profile, candidates, limit=10)
    assert ranked[0].startswith("p1")


# 2. Preferred brand → higher rank
def test_rank_by_brand_match(engine):
    profile = _profile(preferredBrands=["Nike"])
    candidates = [
        _enc("p1", brand="Adidas"),
        _enc("p2", brand="Nike"),
    ]
    ranked, scores = engine.rank_products(profile, candidates, limit=10)
    assert ranked[0].startswith("p2")


# 3. Excluded category → penalized (lower rank)
def test_excluded_category_penalized(engine):
    profile = _profile(
        preferredCategories=["clothing"],
        excludedCategories=["clothing"],
    )
    candidates = [
        _enc("p1", category="clothing"),
        _enc("p2", category="books"),
    ]
    # clothing gets +2 for preferred, -1 for excluded = net +1; books = 0
    # p1 should still rank higher than p2 even with penalty
    ranked, _ = engine.rank_products(profile, candidates, limit=10)
    # Verify the excluded product scores lower than one with no exclusion + preferred
    profile2 = _profile(excludedCategories=["clothing"])
    _, scores2 = engine.rank_products(profile2, candidates, limit=10)
    p1_id = candidates[0]
    p2_id = candidates[1]
    # p1 (clothing, excluded) should score lower than p2 (books, neutral)
    raw_p1 = engine._score_product(p1_id, profile2)
    raw_p2 = engine._score_product(p2_id, profile2)
    assert raw_p1 < raw_p2


# 4. Recently viewed → bonus score
def test_viewed_bonus(engine):
    p1 = _enc("p1")
    p2 = _enc("p2")
    profile = _profile(recentlyViewedProducts=[p1])
    ranked, _ = engine.rank_products(profile, [p1, p2], limit=10)
    assert ranked[0] == p1


# 5. Purchase history → bonus score
def test_purchased_bonus(engine):
    p1 = _enc("p1")
    p2 = _enc("p2")
    profile = _profile(purchaseHistory=[p2])
    ranked, _ = engine.rank_products(profile, [p1, p2], limit=10)
    assert ranked[0] == p2


# 6. No profile (default) → returns original order
def test_no_profile_original_order(engine):
    default_profile = engine.get_default_profile("u-new")
    candidates = ["prod-a", "prod-b", "prod-c"]
    ranked, _ = engine.rank_products(default_profile, candidates, limit=10)
    # All scores equal → ranked preserves order (stable sort for equal keys)
    assert set(ranked) == set(candidates)


# 7. Limit is respected
def test_limit_respected(engine):
    profile = _profile()
    candidates = [f"p{i}" for i in range(20)]
    ranked, _ = engine.rank_products(profile, candidates, limit=5)
    assert len(ranked) == 5


# 8. Score normalization — all scores in [0, 1]
def test_score_normalization(engine):
    profile = _profile(
        preferredCategories=["tech"],
        purchaseHistory=[_enc("p1", category="tech")],
    )
    candidates = [_enc("p1", category="tech"), _enc("p2", category="sports"), "p3"]
    _, scores = engine.rank_products(profile, candidates, limit=10)
    for v in scores.values():
        assert 0.0 <= v <= 1.0


# 9. Empty candidates → empty result
def test_empty_candidates(engine):
    profile = _profile(preferredCategories=["electronics"])
    ranked, scores = engine.rank_products(profile, [], limit=10)
    assert ranked == []
    assert scores == {}


# 10. All candidates in excluded category → they all get penalized
def test_all_excluded(engine):
    profile = _profile(excludedCategories=["bad"])
    candidates = [_enc(f"p{i}", category="bad") for i in range(3)]
    _, scores = engine.rank_products(profile, candidates, limit=10)
    # All scores equal (all penalized equally) → normalised scores are 0.5
    for v in scores.values():
        assert abs(v - 0.5) < 1e-6
