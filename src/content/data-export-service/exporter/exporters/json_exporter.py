"""JSON export implementation.

Converts the tabular (headers + rows) representation into a list of
dictionaries — one dict per row — then serialises with pretty-printing.
"""

import json
from typing import Any, List


class _SafeEncoder(json.JSONEncoder):
    """Fallback encoder that converts non-serialisable objects to strings."""

    def default(self, obj: Any) -> Any:
        try:
            return super().default(obj)
        except TypeError:
            return str(obj)


class JsonExporter:
    def export(self, headers: List[str], rows: List[List[Any]]) -> bytes:
        """Serialise *headers* and *rows* to pretty-printed JSON bytes (UTF-8)."""
        records: List[dict] = []
        for row in rows:
            record: dict = {}
            for idx, header in enumerate(headers):
                value = row[idx] if idx < len(row) else None
                record[header] = value
            records.append(record)

        payload = json.dumps(records, indent=2, ensure_ascii=False, cls=_SafeEncoder)
        return payload.encode("utf-8")
