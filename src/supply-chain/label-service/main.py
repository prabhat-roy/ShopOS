"""ShopOS — label-service entry point."""

import uvicorn
from fastapi import FastAPI
from label.config import get_settings
from label.api import router

settings = get_settings()

app = FastAPI(
    title="label-service",
    description="ShopOS supply-chain label generation service",
    version="1.0.0",
)

app.include_router(router)


if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=settings.http_port,
        reload=False,
        access_log=True,
    )
