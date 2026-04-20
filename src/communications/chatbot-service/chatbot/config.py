from __future__ import annotations

import os
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    HTTP_PORT: int = int(os.getenv("HTTP_PORT", "8193"))
    GRPC_PORT: int = int(os.getenv("GRPC_PORT", "50189"))
    REDIS_URL: str = os.getenv("REDIS_URL", "redis://localhost:6379/0")
    SESSION_TTL_SECONDS: int = int(os.getenv("SESSION_TTL_SECONDS", "1800"))
    LOG_LEVEL: str = os.getenv("LOG_LEVEL", "info")

    class Config:
        env_file = ".env"


settings = Settings()
