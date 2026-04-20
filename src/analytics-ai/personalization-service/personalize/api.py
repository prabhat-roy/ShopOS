from __future__ import annotations

from typing import Annotated, Any

from fastapi import APIRouter, Depends, HTTPException, Request, status

from .models import PersonalizationRequest, PersonalizedResult, UserProfile
from .service import PersonalizationService

router = APIRouter()


def _get_service(request: Request) -> PersonalizationService:
    return request.app.state.service


# ---------------------------------------------------------------------------
# Health
# ---------------------------------------------------------------------------


@router.get("/healthz", tags=["health"])
async def healthz() -> dict:
    return {"status": "ok"}


# ---------------------------------------------------------------------------
# Personalization
# ---------------------------------------------------------------------------


@router.post(
    "/personalize",
    response_model=PersonalizedResult,
    status_code=status.HTTP_200_OK,
    tags=["personalize"],
)
async def personalize(
    req: PersonalizationRequest,
    svc: Annotated[PersonalizationService, Depends(_get_service)],
) -> PersonalizedResult:
    return await svc.personalize(req)


# ---------------------------------------------------------------------------
# Profiles
# ---------------------------------------------------------------------------


@router.get(
    "/profiles/{user_id}",
    response_model=UserProfile,
    tags=["profiles"],
)
async def get_profile(
    user_id: str,
    svc: Annotated[PersonalizationService, Depends(_get_service)],
) -> UserProfile:
    profile = await svc.get_profile(user_id)
    if profile is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Profile for user '{user_id}' not found",
        )
    return profile


@router.put(
    "/profiles/{user_id}",
    response_model=UserProfile,
    status_code=status.HTTP_200_OK,
    tags=["profiles"],
)
async def upsert_profile(
    user_id: str,
    data: dict[str, Any],
    svc: Annotated[PersonalizationService, Depends(_get_service)],
) -> UserProfile:
    return await svc.update_profile(user_id, data)


@router.post(
    "/profiles/{user_id}/view",
    status_code=status.HTTP_204_NO_CONTENT,
    response_model=None,
    tags=["profiles"],
)
async def record_view(
    user_id: str,
    payload: dict[str, str],
    svc: Annotated[PersonalizationService, Depends(_get_service)],
) -> None:
    product_id = payload.get("productId", "")
    if not product_id:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail="productId is required",
        )
    await svc.record_view(user_id, product_id)


@router.post(
    "/profiles/{user_id}/purchase",
    status_code=status.HTTP_204_NO_CONTENT,
    response_model=None,
    tags=["profiles"],
)
async def record_purchase(
    user_id: str,
    payload: dict[str, str],
    svc: Annotated[PersonalizationService, Depends(_get_service)],
) -> None:
    product_id = payload.get("productId", "")
    if not product_id:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail="productId is required",
        )
    await svc.record_purchase(user_id, product_id)


@router.delete(
    "/profiles/{user_id}",
    status_code=status.HTTP_204_NO_CONTENT,
    response_model=None,
    tags=["profiles"],
)
async def delete_profile(
    user_id: str,
    svc: Annotated[PersonalizationService, Depends(_get_service)],
) -> None:
    deleted = await svc.delete_profile(user_id)
    if not deleted:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Profile for user '{user_id}' not found",
        )
