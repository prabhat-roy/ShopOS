from __future__ import annotations

from typing import List
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8", extra="ignore")

    HTTP_PORT: int = 8709

    KAFKA_BROKERS: str = "localhost:9092"
    KAFKA_TOPIC_ANALYTICS: str = "analytics.page.viewed"

    CASSANDRA_HOSTS: str = "127.0.0.1"
    CASSANDRA_KEYSPACE: str = "event_tracking"

    # Store backend: "memory" or "cassandra"
    STORE_BACKEND: str = "cassandra"

    @property
    def cassandra_host_list(self) -> List[str]:
        return [h.strip() for h in self.CASSANDRA_HOSTS.split(",") if h.strip()]


settings = Settings()
