# ============================================================
# Stage 1 — Build
# ============================================================
FROM rust:1.82-alpine AS builder

# Native build dependencies for ring/bcrypt/sqlx
RUN apk add --no-cache musl-dev openssl-dev openssl-libs-static pkgconfig

WORKDIR /app

# Cache dependency compilation separately from source changes
COPY Cargo.toml Cargo.lock* ./
RUN mkdir src && echo 'fn main() {}' > src/main.rs && cargo build --release 2>/dev/null; rm -rf src

# Now copy real source and rebuild (only changed crates recompile)
COPY src ./src
RUN touch src/main.rs && cargo build --release

# ============================================================
# Stage 2 — Runtime
# ============================================================

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /app/target/release/shopos-auth-service /app/shopos-auth-service
USER nonroot:nonroot
EXPOSE 8098 50060
ENTRYPOINT ["/app/shopos-auth-service"]
