import uvicorn
from fastapi import FastAPI
from processor.config import settings
from processor.api import router

app = FastAPI(
    title="Image Processing Service",
    description="Resizes, converts, and optimizes images for the ShopOS platform.",
    version="1.0.0",
)

app.include_router(router)


if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=settings.HTTP_PORT,
        reload=False,
        log_level="info",
    )
