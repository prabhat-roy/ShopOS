from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    http_port: int = 8089
    target_base_url: str = "http://localhost:8080"
    default_rps: float = 10.0
    ramp_up_seconds: int = 30
    scenario: str = "browse"
    duration_seconds: int = 300
    concurrency: int = 10

    class Config:
        env_file = ".env"


settings = Settings()
