from __future__ import annotations

import logging
from typing import List

from fastapi import APIRouter, HTTPException

from .consumer import kafka_consumer
from .models import EmailRecord
from .store import email_store

logger = logging.getLogger(__name__)

router = APIRouter()


# ---------------------------------------------------------------------------
# Health
# ---------------------------------------------------------------------------


@router.get("/healthz", tags=["ops"])
async def healthz():
    """Kubernetes / load-balancer liveness probe."""
    consumer_status = "running" if kafka_consumer.is_running else "stopped"
    return {"status": "ok", "consumer": consumer_status}


# ---------------------------------------------------------------------------
# Individual record lookup
# ---------------------------------------------------------------------------


@router.get("/emails/{message_id}", response_model=EmailRecord, tags=["emails"])
async def get_email(message_id: str):
    """Return the EmailRecord for a given *message_id*."""
    record = await email_store.get(message_id)
    if record is None:
        raise HTTPException(status_code=404, detail=f"Email record not found: {message_id!r}")
    return record


# ---------------------------------------------------------------------------
# List
# ---------------------------------------------------------------------------


@router.get("/emails", response_model=List[EmailRecord], tags=["emails"])
async def list_emails(limit: int = 50):
    """Return up to *limit* most-recently-processed email records (newest first)."""
    if limit < 1 or limit > 500:
        raise HTTPException(status_code=400, detail="limit must be between 1 and 500")
    return await email_store.list(limit=limit)


# ---------------------------------------------------------------------------
# Stats
# ---------------------------------------------------------------------------


@router.get("/email/stats", tags=["emails"])
async def get_stats():
    """Return aggregate delivery statistics.

    Note: the route is intentionally ``/email/stats`` (singular) to avoid
    conflicting with the ``/emails/{message_id}`` wildcard route.
    """
    return await email_store.get_stats()
