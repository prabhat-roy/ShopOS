from __future__ import annotations

import asyncio
import logging
from collections import OrderedDict
from typing import Dict, List, Optional

from .models import EmailRecord

logger = logging.getLogger(__name__)


class EmailStore:
    """Thread-safe in-memory store for EmailRecord objects.

    Keeps at most *max_size* entries; when the limit is reached the oldest
    entry (by insertion order) is evicted before the new one is saved.
    """

    def __init__(self, max_size: int = 10_000) -> None:
        self._max_size = max_size
        self._records: OrderedDict[str, EmailRecord] = OrderedDict()
        self._lock = asyncio.Lock()
        self._total_sent: int = 0
        self._total_delivered: int = 0
        self._total_failed: int = 0

    # ------------------------------------------------------------------
    # Write
    # ------------------------------------------------------------------

    async def save(self, record: EmailRecord) -> None:
        """Persist a record, evicting the oldest entry when at capacity."""
        async with self._lock:
            # Evict oldest entries until we are under the cap
            while len(self._records) >= self._max_size:
                evicted_key, _ = self._records.popitem(last=False)
                logger.debug("Evicted oldest record: %s", evicted_key)

            self._records[record.messageId] = record
            self._total_sent += 1
            if record.status == "delivered":
                self._total_delivered += 1
            else:
                self._total_failed += 1

    # ------------------------------------------------------------------
    # Read
    # ------------------------------------------------------------------

    async def get(self, message_id: str) -> Optional[EmailRecord]:
        """Return the record for *message_id*, or ``None`` if not found."""
        async with self._lock:
            return self._records.get(message_id)

    async def list(self, limit: int = 50) -> List[EmailRecord]:
        """Return up to *limit* most-recently-saved records (newest first)."""
        async with self._lock:
            items = list(self._records.values())
        # Reverse to return newest first without mutating the underlying dict
        return list(reversed(items))[:limit]

    async def get_stats(self) -> Dict[str, int]:
        """Return aggregate delivery statistics."""
        async with self._lock:
            return {
                "sent": self._total_sent,
                "delivered": self._total_delivered,
                "failed": self._total_failed,
            }


# Module-level singleton used by the API and consumer
email_store = EmailStore()
