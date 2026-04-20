from __future__ import annotations

import logging
from typing import Optional

from fastapi import APIRouter, HTTPException, Query

from .engine import recommendation_engine
from .models import (
    InteractionRecordedResponse,
    RecommendRequest,
    RecommendResponse,
    UserInteraction,
)
from .store import interaction_store

logger = logging.getLogger(__name__)

router = APIRouter()


# ------------------------------------------------------------------
# Health
# ------------------------------------------------------------------


@router.get("/healthz", tags=["ops"])
async def healthz() -> dict:
    return {
        "status": "ok",
        "service": "recommendation-service",
        "total_interactions": interaction_store.total_interactions(),
    }


# ------------------------------------------------------------------
# Recommendations
# ------------------------------------------------------------------


@router.post("/recommendations", response_model=RecommendResponse, tags=["recommendations"])
async def get_recommendations(req: RecommendRequest) -> RecommendResponse:
    """
    Return product recommendations based on strategy.

    Strategies: hybrid | user-based | item-based | popular
    """
    try:
        response = recommendation_engine.recommend(req)
    except Exception as exc:
        logger.exception("Error computing recommendations: %s", exc)
        raise HTTPException(status_code=500, detail="Failed to compute recommendations")
    return response


@router.get("/recommendations/popular", response_model=list[dict], tags=["recommendations"])
async def get_popular(
    limit: int = Query(default=10, ge=1, le=100),
) -> list[dict]:
    """Return the globally most popular products."""
    recs = recommendation_engine.recommend_popular(limit=limit)
    return [r.model_dump() for r in recs]


# ------------------------------------------------------------------
# Interactions
# ------------------------------------------------------------------


@router.post(
    "/interactions",
    response_model=InteractionRecordedResponse,
    status_code=201,
    tags=["interactions"],
)
async def record_interaction(interaction: UserInteraction) -> InteractionRecordedResponse:
    """Record a user–product interaction (view, purchase, cart, wishlist)."""
    try:
        interaction_store.add_interaction(interaction)
    except Exception as exc:
        logger.exception("Error recording interaction: %s", exc)
        raise HTTPException(status_code=500, detail="Failed to record interaction")

    return InteractionRecordedResponse(
        recorded=True,
        userId=interaction.userId,
        productId=interaction.productId,
        interactionType=interaction.interactionType,
    )


@router.get(
    "/interactions/user/{userId}",
    response_model=list[UserInteraction],
    tags=["interactions"],
)
async def get_user_interactions(userId: str) -> list[UserInteraction]:
    """Return all recorded interactions for a given user."""
    interactions = interaction_store.get_user_interactions(userId)
    if not interactions:
        raise HTTPException(
            status_code=404,
            detail=f"No interactions found for user '{userId}'",
        )
    return interactions
