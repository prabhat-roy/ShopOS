"""ExportService — routes ExportRequest objects to the correct exporter."""

from typing import Tuple

from exporter.config import settings
from exporter.exporters.csv_exporter import CsvExporter
from exporter.exporters.json_exporter import JsonExporter
from exporter.exporters.pdf_exporter import PdfExporter
from exporter.exporters.xlsx_exporter import XlsxExporter
from exporter.models import ExportFormat, ExportRequest

_CONTENT_TYPES = {
    ExportFormat.CSV: "text/csv; charset=utf-8",
    ExportFormat.JSON: "application/json",
    ExportFormat.XLSX: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
    ExportFormat.PDF: "application/pdf",
}

_EXTENSIONS = {
    ExportFormat.CSV: "csv",
    ExportFormat.JSON: "json",
    ExportFormat.XLSX: "xlsx",
    ExportFormat.PDF: "pdf",
}

_csv_exporter = CsvExporter()
_json_exporter = JsonExporter()
_xlsx_exporter = XlsxExporter()
_pdf_exporter = PdfExporter()


class ExportService:
    def export(self, request: ExportRequest) -> Tuple[bytes, str, str]:
        """
        Generate the export payload for *request*.

        Returns
        -------
        Tuple[bytes, str, str]
            A tuple of (data bytes, MIME content-type, full filename with extension).

        Raises
        ------
        ValueError
            If the number of rows exceeds MAX_ROWS.
        """
        row_count = len(request.rows)
        if row_count > settings.MAX_ROWS:
            raise ValueError(
                f"Row count {row_count} exceeds the maximum allowed limit of {settings.MAX_ROWS} rows."
            )

        fmt = request.format
        headers = request.headers
        rows = request.rows
        ext = _EXTENSIONS[fmt]
        filename = f"{request.filename}.{ext}"
        content_type = _CONTENT_TYPES[fmt]

        if fmt == ExportFormat.CSV:
            data = _csv_exporter.export(headers, rows)
        elif fmt == ExportFormat.JSON:
            data = _json_exporter.export(headers, rows)
        elif fmt == ExportFormat.XLSX:
            data = _xlsx_exporter.export(headers, rows)
        elif fmt == ExportFormat.PDF:
            data = _pdf_exporter.export(headers, rows, title=request.title)
        else:
            raise ValueError(f"Unsupported export format: {fmt}")

        return data, content_type, filename
