# ---- Build stage ----
FROM python:3.12-slim AS builder

WORKDIR /build

RUN apt-get update && apt-get install -y --no-install-recommends \
        gcc \
        libssl-dev \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
RUN pip install --no-cache-dir --prefix=/install -r requirements.txt


# ---- Runtime stage ----
FROM python:3.12-slim AS runtime

LABEL maintainer="ShopOS Platform Team"
LABEL service="search-analytics-service"
LABEL version="1.0.0"

COPY --from=builder /install /usr/local

WORKDIR /app

COPY main.py .
COPY searchanalytics/ searchanalytics/

RUN groupadd -g 10001 appgroup && \
    useradd -u 10001 -g appgroup -M -s /usr/sbin/nologin appuser && \
    chown -R appuser:appgroup /app

USER 10001

EXPOSE 8196

ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1

HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
    CMD python -c "import urllib.request; urllib.request.urlopen('http://localhost:8196/healthz')" || exit 1

CMD ["python", "main.py"]
