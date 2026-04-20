import uvicorn
from fastapi import FastAPI

from generator.config import settings
from handler.api import router

app = FastAPI(
    title="ShopOS Load Generator",
    description="HTTP load generator service for ShopOS platform testing.",
    version="1.0.0",
)

app.include_router(router)


if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=settings.http_port,
        reload=False,
    )
