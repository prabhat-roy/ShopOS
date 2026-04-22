from fastapi import FastAPI
from fastapi.responses import PlainTextResponse
import uvicorn
import os

app = FastAPI(title="etl-service")


@app.get("/healthz")
def healthz():
    return {"status": "ok"}


@app.get("/metrics", response_class=PlainTextResponse)
def metrics():
    return "# placeholder metrics\n"


if __name__ == "__main__":
    port = int(os.getenv("PORT", "8201"))
    uvicorn.run(app, host="0.0.0.0", port=port)
