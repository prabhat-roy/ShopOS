from pydantic_settings import BaseSettings
from typing import List


class Settings(BaseSettings):
    HTTP_PORT: int = 8707
    KAFKA_BROKERS: str = "localhost:9092"
    INPUT_TOPICS: str = "analytics.page.viewed,analytics.product.clicked,analytics.search.performed"
    OUTPUT_TOPIC: str = "analytics.enriched"
    KAFKA_GROUP_ID: str = "data-pipeline"
    CASSANDRA_HOSTS: str = "localhost"
    CASSANDRA_KEYSPACE: str = "pipeline"

    @property
    def input_topics_list(self) -> List[str]:
        return [t.strip() for t in self.INPUT_TOPICS.split(",") if t.strip()]

    @property
    def kafka_brokers_list(self) -> List[str]:
        return [b.strip() for b in self.KAFKA_BROKERS.split(",") if b.strip()]

    @property
    def cassandra_hosts_list(self) -> List[str]:
        return [h.strip() for h in self.CASSANDRA_HOSTS.split(",") if h.strip()]

    class Config:
        env_file = ".env"
        env_file_encoding = "utf-8"


settings = Settings()
