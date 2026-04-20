from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    HTTP_PORT: int = 8703
    SERVICE_NAME: str = "sentiment-analysis-service"
    LOG_LEVEL: str = "INFO"

    DATABASE_URL: str = "postgresql://sentiment:sentiment@localhost:5432/sentimentdb"
    KAFKA_BROKERS: str = "localhost:9092"
    KAFKA_GROUP_ID: str = "sentiment-analysis"
    KAFKA_TOPICS: str = "commerce.order.placed"

    model_config = {"env_file": ".env", "env_file_encoding": "utf-8"}


settings = Settings()
