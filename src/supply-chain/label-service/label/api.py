"""FastAPI router for label-service endpoints."""

from typing import Any
from fastapi import APIRouter, HTTPException, status
from label.models import LabelRequest, LabelResponse
from label.generator import LabelGenerator

router = APIRouter()
_generator = LabelGenerator()


@router.get("/healthz", response_model=dict[str, Any])
def health_check() -> dict[str, Any]:
    """Liveness probe — returns HTTP 200 when the service is running."""
    return {"status": "ok", "service": "label-service"}


@router.post(
    "/labels/generate",
    response_model=LabelResponse,
    status_code=status.HTTP_200_OK,
    summary="Generate a single shipping label",
)
def generate_label(request: LabelRequest) -> LabelResponse:
    """
    Generate a shipping label for a single shipment.

    Supports two label formats:
    - **ZPL** (default): ZPL II code for thermal printers.
    - **TEXT**: Human-readable ASCII label for testing or plain display.
    """
    try:
        return _generator.generate(request)
    except Exception as exc:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Label generation failed: {exc}",
        ) from exc


@router.post(
    "/labels/batch",
    response_model=list[LabelResponse],
    status_code=status.HTTP_200_OK,
    summary="Generate labels for multiple shipments",
)
def generate_labels_batch(requests: list[LabelRequest]) -> list[LabelResponse]:
    """
    Generate shipping labels for a batch of shipments in one call.
    Each item in the request list is processed independently.
    """
    if not requests:
        return []

    results: list[LabelResponse] = []
    errors: list[str] = []

    for i, req in enumerate(requests):
        try:
            results.append(_generator.generate(req))
        except Exception as exc:
            errors.append(f"Item {i} (shipmentId={req.shipment_id}): {exc}")

    if errors:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail={"message": "One or more labels failed to generate", "errors": errors},
        )

    return results
