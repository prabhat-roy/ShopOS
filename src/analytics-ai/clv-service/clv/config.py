from __future__ import annotations

import os
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    HTTP_PORT: int = int(os.getenv("HTTP_PORT", "8195"))
    GRPC_PORT: int = int(os.getenv("GRPC_PORT", "50190"))
    DATABASE_URL: str = os.getenv(
        "DATABASE_URL", "postgresql://clv_user:clv_pass@localhost:5432/clv_db"
    )
    CLV_PREDICTION_HORIZON_DAYS: int = int(os.getenv("CLV_PREDICTION_HORIZON_DAYS", "365"))
    LOG_LEVEL: str = os.getenv("LOG_LEVEL", "info")

    class Config:
        env_file = ".env"


settings = Settings()
