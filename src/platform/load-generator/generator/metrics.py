import statistics
import threading
from collections import deque
from typing import Dict

from .scenarios import ScenarioResult


class MetricsCollector:
    """Thread-safe rolling-window metrics collector."""

    def __init__(self, window: int = 10000) -> None:
        self._lock = threading.Lock()
        self._results: deque = deque(maxlen=window)

    def record(self, result: ScenarioResult) -> None:
        with self._lock:
            self._results.append(result)

    def snapshot(self) -> Dict:
        with self._lock:
            results = list(self._results)

        if not results:
            return {
                "total_requests": 0,
                "success_count": 0,
                "error_count": 0,
                "success_rate": 0.0,
                "avg_latency_ms": 0.0,
                "p95_latency_ms": 0.0,
                "errors_by_status": {},
            }

        total = len(results)
        success_count = sum(1 for r in results if r.success)
        error_count = total - success_count
        success_rate = success_count / total

        latencies = [r.latency_ms for r in results]
        avg_latency_ms = statistics.mean(latencies)

        sorted_latencies = sorted(latencies)
        p95_index = max(0, int(len(sorted_latencies) * 0.95) - 1)
        p95_latency_ms = sorted_latencies[p95_index]

        errors_by_status: Dict[str, int] = {}
        for r in results:
            if not r.success:
                key = str(r.status_code)
                errors_by_status[key] = errors_by_status.get(key, 0) + 1

        return {
            "total_requests": total,
            "success_count": success_count,
            "error_count": error_count,
            "success_rate": success_rate,
            "avg_latency_ms": avg_latency_ms,
            "p95_latency_ms": p95_latency_ms,
            "errors_by_status": errors_by_status,
        }
