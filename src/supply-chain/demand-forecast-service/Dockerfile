# ── Stage 1: builder ──────────────────────────────────────────────────────────
FROM python:3.12-slim AS builder

WORKDIR /app

# Install build deps in an isolated virtual env to keep final image clean
RUN python -m venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"

COPY requirements.txt .
RUN pip install --no-cache-dir --prefix=/install -r requirements.txt

# ── Stage 2: runtime ──────────────────────────────────────────────────────────
FROM python:3.12-slim AS runtime

LABEL org.opencontainers.image.title="demand-forecast-service" \
      org.opencontainers.image.description="Demand forecasting via Kafka + moving average" \
      org.opencontainers.image.version="1.0.0"

# Non-root user (uid 65534 = nobody)
RUN groupadd --gid 10001 appgroup \
 && useradd  --uid 10001 --gid 10001 --no-create-home --shell /usr/sbin/nologin appuser

WORKDIR /app

# Copy virtualenv from builder
COPY --from=builder /opt/venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"

# Copy application source
COPY --chown=appuser:appgroup forecast/ ./forecast/
COPY --chown=appuser:appgroup main.py .

USER 10001

EXPOSE 8209

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD python -c "import urllib.request; urllib.request.urlopen('http://localhost:8209/healthz')"

CMD ["python", "main.py"]
