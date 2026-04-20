from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    http_port: int = 8121
    elasticsearch_url: str = "http://localhost:9200"
    index_name: str = "products"

    class Config:
        env_file = ".env"


settings = Settings()
