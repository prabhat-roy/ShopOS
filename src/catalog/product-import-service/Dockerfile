# ---- Build Stage ----
FROM golang:1.24-alpine AS builder
WORKDIR /app

# Download dependencies first (layer cache)

COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    rm -f go.sum && GONOSUMDB=* go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build \
    rm -f go.sum && GONOSUMDB=* go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build  -trimpath -ldflags="-s -w" -o /product-import-service .



FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /product-import-service /product-import-service
USER nonroot:nonroot
EXPOSE 8120 50180
ENTRYPOINT ["/product-import-service"]
