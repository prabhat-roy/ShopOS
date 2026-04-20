import uvicorn
from fastapi import FastAPI
from exporter.config import settings
from exporter.api import router

app = FastAPI(
    title="Data Export Service",
    description="Exports data in CSV, JSON, XLSX, and PDF formats for the ShopOS platform.",
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
