from pydantic_settings import BaseSettings
from typing import List


class Settings(BaseSettings):
    HTTP_PORT: int = 8601
    GRPC_PORT: int = 50141
    MAX_FILE_SIZE_MB: int = 10
    ALLOWED_FORMATS: List[str] = ["JPEG", "PNG", "WEBP", "GIF", "BMP"]

    model_config = {"env_file": ".env", "env_file_encoding": "utf-8"}


settings = Settings()
