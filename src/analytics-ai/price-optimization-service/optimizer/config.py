from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8", extra="ignore")

    HTTP_PORT: int = 8704
    GRPC_PORT: int = 50151
    DATABASE_URL: str = "postgresql://postgres:postgres@localhost:5432/price_optimization"
    LOG_LEVEL: str = "INFO"
    SERVICE_NAME: str = "price-optimization-service"


settings = Settings()
