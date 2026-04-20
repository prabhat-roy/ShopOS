# ── Stage 1: builder ──────────────────────────────────────────────────────────
FROM python:3.12-slim AS builder

WORKDIR /build

COPY requirements.txt .
RUN pip install --no-cache-dir --prefix=/install -r requirements.txt


# ── Stage 2: runtime ──────────────────────────────────────────────────────────
FROM python:3.12-slim AS runtime

# Copy installed packages from builder
COPY --from=builder /install /usr/local

# Non-root user (uid/gid 65534 = nobody)
RUN groupadd -g 10001 appgroup && useradd -u 10001 -g appgroup -s /sbin/nologin appuser

WORKDIR /app

COPY --chown=appuser:appgroup . .

USER 10001

EXPOSE 8607

CMD ["python", "main.py"]
