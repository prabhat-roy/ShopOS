"""
Unit tests for the RecommendationEngine.

Tests cover:
  1.  Popular with no data → empty list
  2.  Popular after interactions → returns correct ordering
  3.  User recommendations (collaborative filtering)
  4.  Similar products via co-purchase
  5.  Hybrid strategy with userId
  6.  Hybrid with empty userId falls back to popular
  7.  Jaccard similarity correctness
  8.  Limit is respected
  9.  Already-seen products excluded from user-based recs
  10. Strategy='popular' parameter is honoured
"""

from __future__ import annotations

import sys
import os
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from datetime import datetime, timezone

import pytest

from recommender.engine import RecommendationEngine, _jaccard
from recommender.models import RecommendRequest, UserInteraction
from recommender.store import InteractionStore


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


def _store_with_interactions() -> InteractionStore:
    """Return a fresh store pre-populated with deterministic interactions."""
    store = InteractionStore()
    interactions = [
        # user-A bought p1, p2, p3
        UserInteraction(userId="user-A", productId="p1", interactionType="purchase", score=1.0),
        UserInteraction(userId="user-A", productId="p2", interactionType="purchase", score=1.0),
        UserInteraction(userId="user-A", productId="p3", interactionType="purchase", score=1.0),
        # user-B bought p1, p2, p4  (similar to user-A for p1 and p2)
        UserInteraction(userId="user-B", productId="p1", interactionType="purchase", score=1.0),
        UserInteraction(userId="user-B", productId="p2", interactionType="purchase", score=1.0),
        UserInteraction(userId="user-B", productId="p4", interactionType="purchase", score=1.0),
        # user-C bought p5, p6
        UserInteraction(userId="user-C", productId="p5", interactionType="purchase", score=1.0),
        UserInteraction(userId="user-C", productId="p6", interactionType="purchase", score=1.0),
        # Extra views to inflate p1 popularity
        UserInteraction(userId="user-D", productId="p1", interactionType="view", score=1.0),
        UserInteraction(userId="user-D", productId="p1", interactionType="view", score=1.0),
        UserInteraction(userId="user-D", productId="p1", interactionType="view", score=1.0),
    ]
    for i in interactions:
        store.add_interaction(i)
    return store


# ---------------------------------------------------------------------------
# Test 1: popular with no data
# ---------------------------------------------------------------------------


def test_popular_no_data():
    store = InteractionStore()
    engine = RecommendationEngine(store)
    result = engine.recommend_popular(limit=5)
    assert result == [], "Expected empty list when store has no data"


# ---------------------------------------------------------------------------
# Test 2: popular after interactions
# ---------------------------------------------------------------------------


def test_popular_after_interactions():
    store = _store_with_interactions()
    engine = RecommendationEngine(store)

    # p1 has 3 views + 2 purchases → highest cumulative score
    result = engine.recommend_popular(limit=3)
    assert len(result) <= 3
    product_ids = [r.productId for r in result]
    assert "p1" in product_ids, "p1 should be among popular products"
    assert result[0].productId == "p1", "p1 should be the most popular"


# ---------------------------------------------------------------------------
# Test 3: user recommendations (collaborative filtering)
# ---------------------------------------------------------------------------


def test_user_recommendations():
    store = _store_with_interactions()
    engine = RecommendationEngine(store)

    # user-A bought p1,p2,p3.  user-B bought p1,p2,p4.
    # user-A should be recommended p4 (from user-B, their most similar peer)
    recs = engine.recommend_for_user("user-A", limit=5)
    product_ids = [r.productId for r in recs]
    assert "p4" in product_ids, "p4 should be recommended for user-A via user-B similarity"

    # Ensure user-A's own products are NOT in recommendations
    for pid in ("p1", "p2", "p3"):
        assert pid not in product_ids, f"{pid} already seen by user-A, must not appear"


# ---------------------------------------------------------------------------
# Test 4: similar products via co-purchase
# ---------------------------------------------------------------------------


def test_similar_products_co_purchase():
    store = _store_with_interactions()
    engine = RecommendationEngine(store)

    # p1 and p2 are both bought by user-A and user-B → should be co-purchase neighbours
    recs = engine.recommend_similar("p1", limit=5)
    product_ids = [r.productId for r in recs]
    assert "p2" in product_ids, "p2 should be a co-purchase recommendation for p1"
    assert "p1" not in product_ids, "p1 must not recommend itself"


# ---------------------------------------------------------------------------
# Test 5: hybrid strategy with userId
# ---------------------------------------------------------------------------


def test_hybrid_with_user_id():
    store = _store_with_interactions()
    engine = RecommendationEngine(store)

    req = RecommendRequest(userId="user-A", strategy="hybrid", limit=5)
    response = engine.recommend(req)

    assert response.userId == "user-A"
    assert "hybrid" in response.strategy
    assert len(response.recommendations) <= 5


# ---------------------------------------------------------------------------
# Test 6: hybrid with no userId or productId falls back to popular
# ---------------------------------------------------------------------------


def test_hybrid_no_context_falls_back_to_popular():
    store = _store_with_interactions()
    engine = RecommendationEngine(store)

    req = RecommendRequest(strategy="hybrid", limit=5)
    response = engine.recommend(req)

    assert "popular" in response.strategy
    assert len(response.recommendations) > 0


# ---------------------------------------------------------------------------
# Test 7: Jaccard similarity correctness
# ---------------------------------------------------------------------------


def test_jaccard_similarity():
    a = {"p1", "p2", "p3"}
    b = {"p2", "p3", "p4"}
    # intersection = {p2, p3} = 2; union = {p1,p2,p3,p4} = 4
    assert _jaccard(a, b) == pytest.approx(2 / 4)

    assert _jaccard(set(), set()) == 0.0
    assert _jaccard({"x"}, {"x"}) == pytest.approx(1.0)
    assert _jaccard({"x"}, {"y"}) == pytest.approx(0.0)


# ---------------------------------------------------------------------------
# Test 8: limit is respected
# ---------------------------------------------------------------------------


def test_limit_respected():
    store = _store_with_interactions()
    engine = RecommendationEngine(store)

    recs = engine.recommend_popular(limit=2)
    assert len(recs) <= 2

    recs = engine.recommend_for_user("user-A", limit=1)
    assert len(recs) <= 1

    req = RecommendRequest(userId="user-A", strategy="hybrid", limit=1)
    response = engine.recommend(req)
    assert len(response.recommendations) <= 1


# ---------------------------------------------------------------------------
# Test 9: seen products excluded from user-based recommendations
# ---------------------------------------------------------------------------


def test_exclude_seen_products():
    store = _store_with_interactions()
    engine = RecommendationEngine(store)

    user_history = store.get_user_history("user-A")
    recs = engine.recommend_for_user("user-A", limit=10)
    rec_ids = {r.productId for r in recs}

    overlap = rec_ids & user_history
    assert not overlap, f"Seen products appeared in recommendations: {overlap}"


# ---------------------------------------------------------------------------
# Test 10: strategy='popular' parameter is honoured
# ---------------------------------------------------------------------------


def test_strategy_popular_parameter():
    store = _store_with_interactions()
    engine = RecommendationEngine(store)

    req = RecommendRequest(userId="user-A", strategy="popular", limit=5)
    response = engine.recommend(req)

    assert response.strategy == "popular"
    # All recommendations should be in the store's popular list
    popular_ids = set(store.get_popular_products(limit=50))
    for rec in response.recommendations:
        assert rec.productId in popular_ids
