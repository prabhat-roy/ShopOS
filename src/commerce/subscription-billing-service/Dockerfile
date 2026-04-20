# ---- Build Stage ----
FROM golang:1.24-alpine AS builder
WORKDIR /app

# Download dependencies (layer cache)

COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    rm -f go.sum && GONOSUMDB=* go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build \
    rm -f go.sum && GONOSUMDB=* go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build  -ldflags="-s -w" -o subscription-billing-service .



FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /app/subscription-billing-service /subscription-billing-service
USER nonroot:nonroot
EXPOSE 8135 50090
ENTRYPOINT ["/subscription-billing-service"]
