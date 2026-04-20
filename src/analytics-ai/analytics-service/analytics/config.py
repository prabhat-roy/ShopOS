from __future__ import annotations

from typing import List
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8", extra="ignore")

    HTTP_PORT: int = 8700

    KAFKA_BROKERS: str = "localhost:9092"
    KAFKA_TOPICS: str = (
        "analytics.page.viewed,analytics.product.clicked,analytics.search.performed"
    )
    KAFKA_GROUP_ID: str = "analytics-service"

    CASSANDRA_HOSTS: str = "127.0.0.1"
    CASSANDRA_KEYSPACE: str = "analytics"

    # Store backend: "memory" or "cassandra"
    STORE_BACKEND: str = "cassandra"

    @property
    def kafka_topic_list(self) -> List[str]:
        return [t.strip() for t in self.KAFKA_TOPICS.split(",") if t.strip()]

    @property
    def cassandra_host_list(self) -> List[str]:
        return [h.strip() for h in self.CASSANDRA_HOSTS.split(",") if h.strip()]


settings = Settings()
