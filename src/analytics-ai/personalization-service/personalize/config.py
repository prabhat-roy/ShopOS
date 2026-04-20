from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8", extra="ignore")

    HTTP_PORT: int = 8706
    GRPC_PORT: int = 50153
    MONGODB_URI: str = "mongodb://localhost:27017"
    MONGODB_DB: str = "personalization"
    LOG_LEVEL: str = "INFO"
    SERVICE_NAME: str = "personalization-service"


settings = Settings()
