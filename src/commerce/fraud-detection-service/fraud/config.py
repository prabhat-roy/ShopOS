from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    http_port: int = 8143
    database_url: str = "postgresql://postgres:postgres@localhost:5432/fraud_detection"
    risk_threshold_high: int = 60
    risk_threshold_critical: int = 80

    class Config:
        env_file = ".env"


settings = Settings()
