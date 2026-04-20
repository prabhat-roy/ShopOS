# ---- build stage ------------------------------------------------------------
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Download dependencies first (better layer caching).

# Copy source and build a statically-linked binary.
COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    rm -f go.sum && GONOSUMDB=* go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build \
    rm -f go.sum && GONOSUMDB=* go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /api-key-service ./main.go

# ---- final stage ------------------------------------------------------------


FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /api-key-service /api-key-service
USER nonroot:nonroot
EXPOSE 8104 50066
ENTRYPOINT ["/api-key-service"]
