from __future__ import annotations

from typing import Annotated, Optional

from fastapi import APIRouter, Depends, HTTPException, Query, Request, status

from .models import (
    FeatureDefinition,
    FeatureType,
    FeatureValue,
    FeatureVector,
    GetFeaturesRequest,
)
from .store import AsyncPgStore

router = APIRouter()


def _get_store(request: Request) -> AsyncPgStore:
    return request.app.state.store


# ---------------------------------------------------------------------------
# Health
# ---------------------------------------------------------------------------


@router.get("/healthz", tags=["health"])
async def healthz() -> dict:
    return {"status": "ok"}


# ---------------------------------------------------------------------------
# Feature Definitions
# ---------------------------------------------------------------------------


@router.post(
    "/features/definitions",
    response_model=FeatureDefinition,
    status_code=status.HTTP_200_OK,
    tags=["definitions"],
)
async def register_feature(
    definition: FeatureDefinition,
    store: Annotated[AsyncPgStore, Depends(_get_store)],
) -> FeatureDefinition:
    return await store.register_feature(definition)


@router.get(
    "/features/definitions",
    response_model=list[FeatureDefinition],
    tags=["definitions"],
)
async def list_definitions(
    store: Annotated[AsyncPgStore, Depends(_get_store)],
    group: Optional[str] = Query(default=None),
    type: Optional[FeatureType] = Query(default=None),
) -> list[FeatureDefinition]:
    return await store.list_definitions(group=group, feature_type=type)


@router.get(
    "/features/definitions/{group}/{name}",
    response_model=FeatureDefinition,
    tags=["definitions"],
)
async def get_definition(
    group: str,
    name: str,
    store: Annotated[AsyncPgStore, Depends(_get_store)],
) -> FeatureDefinition:
    definition = await store.get_definition(name, group)
    if definition is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Feature definition '{name}' not found in group '{group}'",
        )
    return definition


# ---------------------------------------------------------------------------
# Feature Values
# ---------------------------------------------------------------------------


@router.post(
    "/features/values",
    status_code=status.HTTP_204_NO_CONTENT,
    response_model=None,
    tags=["values"],
)
async def save_value(
    fv: FeatureValue,
    store: Annotated[AsyncPgStore, Depends(_get_store)],
) -> None:
    await store.save_value(fv)


@router.post(
    "/features/values/batch",
    status_code=status.HTTP_204_NO_CONTENT,
    response_model=None,
    tags=["values"],
)
async def save_batch(
    values: list[FeatureValue],
    store: Annotated[AsyncPgStore, Depends(_get_store)],
) -> None:
    await store.save_batch(values)


# ---------------------------------------------------------------------------
# Feature Vector retrieval
# ---------------------------------------------------------------------------


@router.get(
    "/features/vector",
    response_model=FeatureVector,
    tags=["vector"],
)
async def get_feature_vector(
    store: Annotated[AsyncPgStore, Depends(_get_store)],
    entityId: str = Query(...),
    group: str = Query(...),
    features: str = Query(..., description="Comma-separated feature names"),
) -> FeatureVector:
    feature_names = [f.strip() for f in features.split(",") if f.strip()]
    if not feature_names:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail="At least one feature name is required",
        )
    return await store.get_feature_vector(entityId, feature_names, group)


# ---------------------------------------------------------------------------
# Entity features
# ---------------------------------------------------------------------------


@router.get(
    "/features/entity/{entity_id}",
    response_model=list[FeatureValue],
    tags=["values"],
)
async def get_entity_features(
    entity_id: str,
    store: Annotated[AsyncPgStore, Depends(_get_store)],
    group: str = Query(...),
) -> list[FeatureValue]:
    return await store.get_entity_features(entity_id, group)


@router.delete(
    "/features/entity/{entity_id}",
    status_code=status.HTTP_204_NO_CONTENT,
    response_model=None,
    tags=["values"],
)
async def delete_entity_features(
    entity_id: str,
    store: Annotated[AsyncPgStore, Depends(_get_store)],
    group: str = Query(...),
) -> None:
    await store.delete_entity_features(entity_id, group)
