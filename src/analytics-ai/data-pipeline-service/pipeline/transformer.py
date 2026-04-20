import time
import uuid
from datetime import datetime, timezone
from typing import Any, Dict, List, Optional

from pipeline.models import EnrichedEvent, RawEvent, TransformRule


_GEO_PREFIXES: Dict[str, str] = {
    "10.": "internal",
    "192.168.": "local",
    "172.16.": "private",
    "172.17.": "private",
    "172.18.": "private",
    "127.": "loopback",
}

_PLATFORM_SIGNATURES: Dict[str, str] = {
    "android": "android",
    "iphone": "ios",
    "ipad": "ios",
    "windows phone": "windows-mobile",
    "windows": "windows",
    "macintosh": "macos",
    "linux": "linux",
    "curl": "cli",
    "python-requests": "cli",
    "go-http-client": "cli",
}


def _detect_platform(user_agent: Optional[str]) -> str:
    if not user_agent:
        return "unknown"
    ua_lower = user_agent.lower()
    for signature, platform in _PLATFORM_SIGNATURES.items():
        if signature in ua_lower:
            return platform
    return "other"


def _detect_geo_region(ip_address: Optional[str]) -> str:
    if not ip_address:
        return "external"
    for prefix, region in _GEO_PREFIXES.items():
        if ip_address.startswith(prefix):
            return region
    return "external"


def _normalize_timestamp(ts: Optional[str]) -> Optional[str]:
    if not ts:
        return datetime.now(timezone.utc).isoformat()
    for fmt in (
        "%Y-%m-%dT%H:%M:%S.%fZ",
        "%Y-%m-%dT%H:%M:%SZ",
        "%Y-%m-%dT%H:%M:%S",
        "%Y-%m-%d %H:%M:%S",
        "%Y-%m-%d",
    ):
        try:
            parsed = datetime.strptime(ts, fmt)
            if parsed.tzinfo is None:
                parsed = parsed.replace(tzinfo=timezone.utc)
            return parsed.isoformat()
        except ValueError:
            continue
    return ts


def _apply_rule(data: Dict[str, Any], rule: TransformRule) -> Dict[str, Any]:
    result = dict(data)
    field  = rule.field

    if rule.action == "rename":
        new_name = rule.params.get("to")
        if new_name and field in result:
            result[new_name] = result.pop(field)

    elif rule.action == "drop":
        result.pop(field, None)

    elif rule.action == "uppercase":
        if field in result and isinstance(result[field], str):
            result[field] = result[field].upper()

    elif rule.action == "lowercase":
        if field in result and isinstance(result[field], str):
            result[field] = result[field].lower()

    elif rule.action == "default":
        default_value = rule.params.get("value")
        if field not in result or result[field] is None:
            result[field] = default_value

    return result


class EventTransformer:
    def __init__(self, rules: Optional[List[TransformRule]] = None) -> None:
        self._rules = rules or []

    def transform(self, raw_event: RawEvent) -> EnrichedEvent:
        start_ns = time.perf_counter_ns()

        event_id = raw_event.eventId or raw_event.data.get("eventId") or str(uuid.uuid4())

        enriched: Dict[str, Any] = dict(raw_event.data)

        enriched["eventId"] = event_id

        raw_ts = enriched.get("timestamp")
        enriched["timestamp"] = _normalize_timestamp(raw_ts)

        user_agent = enriched.get("userAgent") or enriched.get("user_agent")
        enriched["platform"] = _detect_platform(user_agent)

        ip_address = enriched.get("ipAddress") or enriched.get("ip_address") or enriched.get("ip")
        enriched["geo_region"] = _detect_geo_region(ip_address)

        enriched.setdefault("session_duration", 0)

        for rule in self._rules:
            enriched = _apply_rule(enriched, rule)

        elapsed_ns = time.perf_counter_ns() - start_ns
        processing_ms = elapsed_ns / 1_000_000.0

        return EnrichedEvent(
            eventId=event_id,
            topic=raw_event.topic,
            originalData=dict(raw_event.data),
            enrichedData=enriched,
            transformedAt=datetime.now(timezone.utc),
            processingTimeMs=processing_ms,
        )
