import asyncio
import time
from typing import Callable, Optional

import httpx

from .metrics import MetricsCollector
from .scenarios import (
    SCENARIO_MAP,
    browse_scenario,
    mixed_scenario,
    purchase_scenario,
    search_scenario,
)


class LoadRunner:
    """Drives concurrent load against a target service."""

    def __init__(self) -> None:
        self._running: bool = False
        self._task: Optional[asyncio.Task] = None
        self.metrics: MetricsCollector = MetricsCollector()
        self._scenario: str = "browse"
        self._started_at: Optional[float] = None

    # ------------------------------------------------------------------
    # Public API
    # ------------------------------------------------------------------

    async def start(
        self,
        rps: float,
        scenario: str,
        duration_seconds: int,
        concurrency: int,
        base_url: str,
    ) -> None:
        """Start the load test.  No-op if already running."""
        if self._running:
            return

        scenario_fn = SCENARIO_MAP.get(scenario, browse_scenario)
        self._scenario = scenario
        self._running = True
        self._started_at = time.monotonic()
        self.metrics = MetricsCollector()  # fresh metrics for each run

        self._task = asyncio.create_task(
            self._run(rps, scenario_fn, duration_seconds, concurrency, base_url)
        )

    async def stop(self) -> None:
        """Stop the load test gracefully."""
        self._running = False
        if self._task and not self._task.done():
            self._task.cancel()
            try:
                await self._task
            except asyncio.CancelledError:
                pass
        self._task = None

    def status(self) -> dict:
        elapsed = (
            time.monotonic() - self._started_at if self._started_at is not None else 0.0
        )
        return {
            "running": self._running,
            "scenario": self._scenario,
            "elapsed_seconds": round(elapsed, 1),
        }

    # ------------------------------------------------------------------
    # Internal
    # ------------------------------------------------------------------

    async def _run(
        self,
        rps: float,
        scenario_fn: Callable,
        duration: int,
        concurrency: int,
        base_url: str,
    ) -> None:
        semaphore = asyncio.Semaphore(concurrency)
        interval = 1.0 / rps if rps > 0 else 0.1
        deadline = time.monotonic() + duration

        async with httpx.AsyncClient(timeout=10.0) as client:
            while self._running and time.monotonic() < deadline:
                async def _fire() -> None:
                    async with semaphore:
                        result = await scenario_fn(client, base_url)
                        self.metrics.record(result)

                asyncio.create_task(_fire())
                await asyncio.sleep(interval)

        self._running = False
