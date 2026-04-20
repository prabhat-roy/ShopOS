from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    HTTP_PORT: int = 8702
    GRPC_PORT: int = 50150
    MAX_RECOMMENDATIONS: int = 20
    SERVICE_NAME: str = "recommendation-service"
    LOG_LEVEL: str = "INFO"

    model_config = {"env_file": ".env", "env_file_encoding": "utf-8"}


settings = Settings()
