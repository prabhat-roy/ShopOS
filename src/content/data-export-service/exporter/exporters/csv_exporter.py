"""CSV export implementation.

Writes a UTF-8 BOM at the start of the file so that Microsoft Excel opens
the file correctly without needing an import wizard.
"""

import csv
import io
from typing import Any, List


class CsvExporter:
    # UTF-8 BOM: makes Excel auto-detect encoding on Windows
    _BOM = "\ufeff"

    def export(self, headers: List[str], rows: List[List[Any]]) -> bytes:
        """Serialise *headers* and *rows* to CSV bytes (UTF-8 with BOM)."""
        buf = io.StringIO()
        buf.write(self._BOM)

        writer = csv.writer(
            buf,
            dialect="excel",
            quoting=csv.QUOTE_MINIMAL,
        )
        writer.writerow(headers)

        for row in rows:
            # Normalise each cell: None → empty string, everything else → str
            normalised = [
                "" if cell is None else str(cell)
                for cell in row
            ]
            writer.writerow(normalised)

        return buf.getvalue().encode("utf-8")
