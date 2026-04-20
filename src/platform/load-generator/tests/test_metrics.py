import pytest

from generator.metrics import MetricsCollector
from generator.scenarios import ScenarioResult


def _make_result(success: bool, status_code: int = 200, latency_ms: float = 50.0) -> ScenarioResult:
    return ScenarioResult(
        scenario_name="browse",
        status_code=status_code,
        latency_ms=latency_ms,
        success=success,
    )


# ---------------------------------------------------------------------------
# Empty collector
# ---------------------------------------------------------------------------

def test_snapshot_empty():
    collector = MetricsCollector()
    snap = collector.snapshot()

    assert snap["total_requests"] == 0
    assert snap["success_count"] == 0
    assert snap["error_count"] == 0
    assert snap["success_rate"] == 0.0
    assert snap["avg_latency_ms"] == 0.0
    assert snap["p95_latency_ms"] == 0.0
    assert snap["errors_by_status"] == {}


# ---------------------------------------------------------------------------
# 10 recorded results (8 success, 2 failures)
# ---------------------------------------------------------------------------

def test_snapshot_ten_results():
    collector = MetricsCollector()

    for i in range(8):
        collector.record(_make_result(success=True, latency_ms=float(10 * (i + 1))))

    collector.record(_make_result(success=False, status_code=500, latency_ms=200.0))
    collector.record(_make_result(success=False, status_code=503, latency_ms=300.0))

    snap = collector.snapshot()

    assert snap["total_requests"] == 10
    assert snap["success_count"] == 8
    assert snap["error_count"] == 2
    assert 0.0 <= snap["success_rate"] <= 1.0
    assert snap["success_rate"] == pytest.approx(0.8)
    assert snap["avg_latency_ms"] > 0
    assert snap["p95_latency_ms"] > 0
    assert "500" in snap["errors_by_status"]
    assert "503" in snap["errors_by_status"]
    assert snap["errors_by_status"]["500"] == 1
    assert snap["errors_by_status"]["503"] == 1


# ---------------------------------------------------------------------------
# All successes — no errors_by_status entries
# ---------------------------------------------------------------------------

def test_snapshot_all_success():
    collector = MetricsCollector()
    for _ in range(5):
        collector.record(_make_result(success=True))

    snap = collector.snapshot()
    assert snap["success_rate"] == 1.0
    assert snap["error_count"] == 0
    assert snap["errors_by_status"] == {}


# ---------------------------------------------------------------------------
# Rolling window eviction
# ---------------------------------------------------------------------------

def test_window_eviction():
    collector = MetricsCollector(window=3)
    for i in range(5):
        collector.record(_make_result(success=True, latency_ms=float(i + 1)))

    snap = collector.snapshot()
    # Only the last 3 records should remain.
    assert snap["total_requests"] == 3


# ---------------------------------------------------------------------------
# Success rate is between 0 and 1
# ---------------------------------------------------------------------------

def test_success_rate_bounds():
    collector = MetricsCollector()
    collector.record(_make_result(success=True))
    collector.record(_make_result(success=False, status_code=400))

    snap = collector.snapshot()
    assert 0.0 <= snap["success_rate"] <= 1.0
