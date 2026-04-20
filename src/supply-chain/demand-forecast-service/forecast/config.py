from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8", extra="ignore")

    HTTP_PORT: int = 8209
    DATABASE_URL: str = "postgresql://postgres:password@localhost:5432/demand_forecast"
    KAFKA_BROKERS: str = "localhost:9092"
    KAFKA_TOPIC: str = "commerce.order.placed"
    KAFKA_GROUP_ID: str = "demand-forecast-consumer"


settings = Settings()
