from typing import Optional

from fastapi import APIRouter, HTTPException
from pydantic import BaseModel, Field

from generator.metrics import MetricsCollector
from generator.runner import LoadRunner

router = APIRouter()

# Single shared runner instance used by the whole process.
_runner = LoadRunner()


# ------------------------------------------------------------------
# Request / response models
# ------------------------------------------------------------------

class StartRequest(BaseModel):
    rps: float = Field(default=10.0, gt=0, description="Requests per second")
    scenario: str = Field(default="browse", description="browse | search | purchase | mixed")
    duration_seconds: int = Field(default=300, gt=0, description="Test duration in seconds")
    concurrency: int = Field(default=10, gt=0, description="Max concurrent requests")
    base_url: Optional[str] = Field(default=None, description="Override target base URL")


# ------------------------------------------------------------------
# Routes
# ------------------------------------------------------------------

@router.get("/healthz")
async def healthz():
    return {"status": "ok"}


@router.get("/metrics")
async def metrics():
    return _runner.metrics.snapshot()


@router.post("/start")
async def start(req: StartRequest):
    from generator.config import settings

    if _runner.status()["running"]:
        raise HTTPException(status_code=409, detail="Load test already running")

    base_url = req.base_url or settings.target_base_url
    await _runner.start(
        rps=req.rps,
        scenario=req.scenario,
        duration_seconds=req.duration_seconds,
        concurrency=req.concurrency,
        base_url=base_url,
    )
    return {"started": True, "scenario": req.scenario, "rps": req.rps}


@router.post("/stop")
async def stop():
    await _runner.stop()
    return {"stopped": True}


@router.get("/status")
async def status():
    return _runner.status()
