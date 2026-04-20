# ── Stage 1: Build ──────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src

# Download dependencies first (layer cache)

# Copy source and build a statically-linked binary
COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    rm -f go.sum && GONOSUMDB=* go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build \
    rm -f go.sum && GONOSUMDB=* go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /bin/product-catalog-service .

# ── Stage 2: Runtime ────────────────────────────────────────────────────────


FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /bin/product-catalog-service /product-catalog-service
USER nonroot:nonroot
EXPOSE 8110 50070
ENTRYPOINT ["/product-catalog-service"]
