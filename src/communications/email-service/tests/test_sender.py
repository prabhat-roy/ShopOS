"""Tests for EmailSender."""
from __future__ import annotations

import pytest

from app.models import EmailMessage, EmailRecord
from app.sender import EmailSender, _validate_email


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


def make_msg(**overrides) -> EmailMessage:
    defaults = dict(
        messageId="msg-001",
        to="recipient@example.com",
        subject="Hello",
        body="Plain text body",
        htmlBody="<p>HTML body</p>",
    )
    defaults.update(overrides)
    return EmailMessage(**defaults)


# Always succeeds (success_rate=1.0 means random() < 1.0 is always True)
SUCCESS_SENDER = EmailSender(success_rate=1.0)
# Always fails (success_rate=0.0 means random() < 0.0 is never True)
FAILURE_SENDER = EmailSender(success_rate=0.0)


# ---------------------------------------------------------------------------
# Tests
# ---------------------------------------------------------------------------


class TestValidateEmail:
    def test_valid_simple_address(self):
        assert _validate_email("user@example.com") is True

    def test_valid_subdomain(self):
        assert _validate_email("user@mail.example.co.uk") is True

    def test_invalid_missing_at(self):
        assert _validate_email("userexample.com") is False

    def test_invalid_missing_dot(self):
        assert _validate_email("user@examplecom") is False

    def test_invalid_empty_string(self):
        assert _validate_email("") is False

    def test_invalid_only_at(self):
        assert _validate_email("@") is False


class TestEmailSenderSuccess:
    def test_successful_send_returns_delivered_record(self):
        msg = make_msg()
        record = SUCCESS_SENDER.send(msg)
        assert isinstance(record, EmailRecord)
        assert record.status == "delivered"
        assert record.messageId == "msg-001"
        assert record.to == "recipient@example.com"
        assert record.errorMessage is None

    def test_successful_send_records_subject(self):
        msg = make_msg(subject="Order Confirmation #42")
        record = SUCCESS_SENDER.send(msg)
        assert record.subject == "Order Confirmation #42"

    def test_sent_at_is_set(self):
        msg = make_msg()
        record = SUCCESS_SENDER.send(msg)
        assert record.sentAt is not None

    def test_html_body_message_still_processes(self):
        msg = make_msg(htmlBody="<h1>Welcome</h1>")
        record = SUCCESS_SENDER.send(msg)
        assert record.status == "delivered"


class TestEmailSenderFailure:
    def test_simulated_failure_returns_failed_record(self):
        msg = make_msg()
        record = FAILURE_SENDER.send(msg)
        assert record.status == "failed"
        assert record.errorMessage is not None
        assert "SMTP" in record.errorMessage or "error" in record.errorMessage.lower()

    def test_invalid_email_returns_failed_record(self):
        msg = make_msg(to="not-an-email")
        record = SUCCESS_SENDER.send(msg)
        assert record.status == "failed"
        assert "Invalid" in (record.errorMessage or "")

    def test_plaintext_only_message_processes(self):
        msg = make_msg(htmlBody=None)
        record = SUCCESS_SENDER.send(msg)
        assert record.status == "delivered"

    def test_message_with_metadata(self):
        msg = make_msg(metadata={"orderId": "ORD-999", "userId": "USR-1"})
        record = SUCCESS_SENDER.send(msg)
        # metadata is passed through but not stored on record
        assert record.messageId == "msg-001"
