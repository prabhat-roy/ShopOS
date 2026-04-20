from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8", extra="ignore")

    HTTP_PORT: int = 8503
    KAFKA_BROKERS: str = "localhost:9092"
    KAFKA_GROUP_ID: str = "email-service"
    KAFKA_TOPIC: str = "email.send"

    # Consumer behaviour
    KAFKA_AUTO_OFFSET_RESET: str = "earliest"
    KAFKA_SESSION_TIMEOUT_MS: int = 30000
    KAFKA_HEARTBEAT_INTERVAL_MS: int = 3000

    # Store limits
    MAX_STORE_SIZE: int = 10000


settings = Settings()
