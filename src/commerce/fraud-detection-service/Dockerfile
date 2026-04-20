# ── Build stage ──────────────────────────────────────────────────────────────
FROM python:3.12-slim AS builder

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir --prefix=/install -r requirements.txt

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM python:3.12-slim

# Non-root user
RUN addgroup --gid 10001 appgroup && \
    adduser --uid 10001 --gid 10001 --no-create-home --disabled-password appuser

WORKDIR /app

COPY --from=builder /install /usr/local
COPY . .

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

EXPOSE 8143

USER 10001

CMD ["python", "-m", "uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8143"]
