import pytest
from datetime import datetime, timezone

from pipeline.models import RawEvent, TransformRule
from pipeline.transformer import EventTransformer


def make_raw(data: dict, topic: str = "analytics.page.viewed", event_id: str = None) -> RawEvent:
    return RawEvent(topic=topic, eventId=event_id, data=data)


def test_adds_event_id_when_missing():
    transformer = EventTransformer()
    raw = make_raw({"sessionId": "s-1", "pageUrl": "/home", "timestamp": "2024-03-10T10:00:00Z"})
    result = transformer.transform(raw)
    assert result.eventId
    assert len(result.eventId) == 36  # UUID format


def test_preserves_event_id_when_present():
    transformer = EventTransformer()
    raw = make_raw({"eventId": "my-custom-id", "timestamp": "2024-03-10T10:00:00Z"})
    result = transformer.transform(raw)
    assert result.eventId == "my-custom-id"


def test_normalizes_timestamp_to_iso():
    transformer = EventTransformer()
    raw = make_raw({"timestamp": "2024-03-10T10:00:00Z"})
    result = transformer.transform(raw)
    ts = result.enrichedData["timestamp"]
    assert "T" in ts
    parsed = datetime.fromisoformat(ts)
    assert parsed is not None


def test_adds_timestamp_when_missing():
    transformer = EventTransformer()
    raw = make_raw({"sessionId": "abc"})
    result = transformer.transform(raw)
    assert "timestamp" in result.enrichedData
    assert result.enrichedData["timestamp"] is not None


def test_platform_detection_android():
    transformer = EventTransformer()
    raw = make_raw({"userAgent": "Mozilla/5.0 (Linux; Android 12; Pixel 6)"})
    result = transformer.transform(raw)
    assert result.enrichedData["platform"] == "android"


def test_platform_detection_ios():
    transformer = EventTransformer()
    raw = make_raw({"userAgent": "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0)"})
    result = transformer.transform(raw)
    assert result.enrichedData["platform"] == "ios"


def test_platform_unknown_when_no_user_agent():
    transformer = EventTransformer()
    raw = make_raw({"sessionId": "s-1"})
    result = transformer.transform(raw)
    assert result.enrichedData["platform"] == "unknown"


def test_geo_region_internal_ip():
    transformer = EventTransformer()
    raw = make_raw({"ipAddress": "10.0.1.25"})
    result = transformer.transform(raw)
    assert result.enrichedData["geo_region"] == "internal"


def test_geo_region_local_ip():
    transformer = EventTransformer()
    raw = make_raw({"ipAddress": "192.168.1.100"})
    result = transformer.transform(raw)
    assert result.enrichedData["geo_region"] == "local"


def test_geo_region_external_ip():
    transformer = EventTransformer()
    raw = make_raw({"ipAddress": "8.8.8.8"})
    result = transformer.transform(raw)
    assert result.enrichedData["geo_region"] == "external"


def test_geo_region_external_when_no_ip():
    transformer = EventTransformer()
    raw = make_raw({"sessionId": "s-1"})
    result = transformer.transform(raw)
    assert result.enrichedData["geo_region"] == "external"


def test_session_duration_default_zero():
    transformer = EventTransformer()
    raw = make_raw({"sessionId": "s-1"})
    result = transformer.transform(raw)
    assert result.enrichedData["session_duration"] == 0


def test_processing_time_is_positive():
    transformer = EventTransformer()
    raw = make_raw({"sessionId": "s-1"})
    result = transformer.transform(raw)
    assert result.processingTimeMs >= 0.0


def test_enriched_data_contains_required_fields():
    transformer = EventTransformer()
    raw = make_raw({
        "sessionId": "s-1",
        "ipAddress": "10.1.2.3",
        "userAgent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
        "timestamp": "2024-04-01T08:30:00Z",
    })
    result = transformer.transform(raw)
    for field in ("eventId", "platform", "geo_region", "session_duration", "timestamp"):
        assert field in result.enrichedData, f"Missing field: {field}"


def test_transform_rule_drop():
    rule = TransformRule(field="password", action="drop", params={})
    transformer = EventTransformer(rules=[rule])
    raw = make_raw({"sessionId": "s-1", "password": "secret"})
    result = transformer.transform(raw)
    assert "password" not in result.enrichedData


def test_transform_rule_uppercase():
    rule = TransformRule(field="country", action="uppercase", params={})
    transformer = EventTransformer(rules=[rule])
    raw = make_raw({"country": "us"})
    result = transformer.transform(raw)
    assert result.enrichedData["country"] == "US"


def test_transform_rule_default():
    rule = TransformRule(field="currency", action="default", params={"value": "USD"})
    transformer = EventTransformer(rules=[rule])
    raw = make_raw({"sessionId": "s-1"})
    result = transformer.transform(raw)
    assert result.enrichedData["currency"] == "USD"


def test_original_data_unchanged():
    transformer = EventTransformer()
    data = {"sessionId": "s-1", "ipAddress": "10.0.0.1", "timestamp": "2024-03-10T10:00:00Z"}
    raw = make_raw(data)
    result = transformer.transform(raw)
    assert result.originalData == data
