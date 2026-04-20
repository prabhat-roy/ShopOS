from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    HTTP_PORT: int = 8607
    GRPC_PORT: int = 50147
    MAX_ROWS: int = 100000

    model_config = {"env_file": ".env", "env_file_encoding": "utf-8"}


settings = Settings()
