"""FastAPI router for the Data Export Service."""

from typing import Dict

from fastapi import APIRouter, HTTPException, status
from fastapi.responses import Response

from exporter.models import ExportFormat, ExportRequest
from exporter.service import ExportService

router = APIRouter()
_service = ExportService()


def _run_export(request: ExportRequest) -> Response:
    """Execute the export and return an HTTP file-download response."""
    try:
        data, content_type, filename = _service.export(request)
    except ValueError as exc:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(exc),
        )
    return Response(
        content=data,
        media_type=content_type,
        headers={"Content-Disposition": f'attachment; filename="{filename}"'},
    )


@router.get("/healthz", tags=["health"])
async def healthz() -> Dict[str, str]:
    return {"status": "ok"}


@router.post(
    "/export",
    tags=["export"],
    summary="Export data in the requested format (CSV, JSON, XLSX, PDF)",
    response_description="Binary file download",
)
async def export_data(request: ExportRequest) -> Response:
    return _run_export(request)


@router.post(
    "/export/csv",
    tags=["export"],
    summary="Export data as CSV",
    response_description="CSV file download",
)
async def export_csv(request: ExportRequest) -> Response:
    return _run_export(request.model_copy(update={"format": ExportFormat.CSV}))


@router.post(
    "/export/json",
    tags=["export"],
    summary="Export data as JSON",
    response_description="JSON file download",
)
async def export_json(request: ExportRequest) -> Response:
    return _run_export(request.model_copy(update={"format": ExportFormat.JSON}))


@router.post(
    "/export/xlsx",
    tags=["export"],
    summary="Export data as Excel XLSX",
    response_description="XLSX file download",
)
async def export_xlsx(request: ExportRequest) -> Response:
    return _run_export(request.model_copy(update={"format": ExportFormat.XLSX}))


@router.post(
    "/export/pdf",
    tags=["export"],
    summary="Export data as PDF",
    response_description="PDF file download",
)
async def export_pdf(request: ExportRequest) -> Response:
    return _run_export(request.model_copy(update={"format": ExportFormat.PDF}))
