from __future__ import annotations

import os
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    HTTP_PORT: int = int(os.getenv("HTTP_PORT", "8196"))
    KAFKA_BROKERS: str = os.getenv("KAFKA_BROKERS", "localhost:9092")
    KAFKA_GROUP_ID: str = os.getenv("KAFKA_GROUP_ID", "search-analytics-service")
    KAFKA_TOPICS: str = os.getenv("KAFKA_TOPICS", "analytics.search.performed")
    CASSANDRA_CONTACT_POINTS: str = os.getenv("CASSANDRA_CONTACT_POINTS", "localhost")
    CASSANDRA_KEYSPACE: str = os.getenv("CASSANDRA_KEYSPACE", "search_analytics")
    LOG_LEVEL: str = os.getenv("LOG_LEVEL", "info")

    class Config:
        env_file = ".env"


settings = Settings()
