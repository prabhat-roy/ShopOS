from __future__ import annotations

import logging
import random
import re
from datetime import datetime, timezone
from email.mime.multipart import MIMEMultipart  # stdlib email — not our package
from email.mime.text import MIMEText            # stdlib email — not our package

from .models import EmailMessage, EmailRecord

logger = logging.getLogger(__name__)

# A deliberately simple regex that catches obviously malformed addresses without
# pulling in a heavy validation library.  Full RFC-5322 compliance is handled by
# the upstream producers / validation service.
_EMAIL_RE = re.compile(r"^[^@\s]+@[^@\s]+\.[^@\s]+$")

# Simulated delivery success rate (95 %)
_SUCCESS_RATE: float = 0.95


def _validate_email(address: str) -> bool:
    """Return True when *address* looks like a valid email address."""
    return bool(_EMAIL_RE.match(address.strip()))


def _build_mime_message(msg: EmailMessage) -> MIMEMultipart:
    """Construct a MIME email object from *msg*.

    The message is never actually transmitted; building it exercises the
    standard library and validates that the content is well-formed.
    """
    mime = MIMEMultipart("alternative")
    mime["Subject"] = msg.subject
    mime["From"] = msg.from_addr
    mime["To"] = msg.to
    mime["Message-ID"] = msg.messageId

    # Plain-text part
    mime.attach(MIMEText(msg.body, "plain", "utf-8"))

    # Optional HTML part
    if msg.htmlBody:
        mime.attach(MIMEText(msg.htmlBody, "html", "utf-8"))

    return mime


class EmailSender:
    """Simulates an SMTP email sender.

    ``send`` constructs a MIME message, validates the recipient address, then
    uses a random draw to simulate a 95 % delivery success rate.  No network
    connection is ever opened.
    """

    def __init__(self, success_rate: float = _SUCCESS_RATE, seed: int | None = None) -> None:
        self._success_rate = success_rate
        self._rng = random.Random(seed)

    def send(self, msg: EmailMessage) -> EmailRecord:
        """Attempt delivery of *msg* and return an EmailRecord."""
        now = datetime.now(tz=timezone.utc)

        if not _validate_email(msg.to):
            logger.warning("Invalid recipient address: %r", msg.to)
            return EmailRecord(
                messageId=msg.messageId,
                to=msg.to,
                subject=msg.subject,
                status="failed",
                sentAt=now,
                errorMessage=f"Invalid recipient email address: {msg.to!r}",
            )

        # Build MIME object (validates content structure)
        try:
            mime_msg = _build_mime_message(msg)
        except Exception as exc:  # pragma: no cover
            logger.error("MIME construction failed for %s: %s", msg.messageId, exc)
            return EmailRecord(
                messageId=msg.messageId,
                to=msg.to,
                subject=msg.subject,
                status="failed",
                sentAt=now,
                errorMessage=f"MIME construction error: {exc}",
            )

        logger.debug(
            "MIME message constructed — From: %s, To: %s, Subject: %s, Size: %d bytes",
            mime_msg["From"],
            mime_msg["To"],
            mime_msg["Subject"],
            len(mime_msg.as_bytes()),
        )

        # Simulate SMTP delivery
        if self._rng.random() < self._success_rate:
            logger.info("Simulated delivery SUCCESS for messageId=%s to=%s", msg.messageId, msg.to)
            return EmailRecord(
                messageId=msg.messageId,
                to=msg.to,
                subject=msg.subject,
                status="delivered",
                sentAt=now,
            )
        else:
            error = "Simulated SMTP relay error: connection timeout"
            logger.warning("Simulated delivery FAILURE for messageId=%s: %s", msg.messageId, error)
            return EmailRecord(
                messageId=msg.messageId,
                to=msg.to,
                subject=msg.subject,
                status="failed",
                sentAt=now,
                errorMessage=error,
            )


# Module-level singleton
email_sender = EmailSender()
