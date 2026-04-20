# ---- build stage ----
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git



COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    rm -f go.sum && GONOSUMDB=* go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build \
    rm -f go.sum && GONOSUMDB=* go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build  -trimpath -ldflags="-s -w" -o /video-service .

# ---- runtime stage ----
# Use alpine so CA certificates are available for MinIO TLS connections.


FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /video-service /video-service
USER nonroot:nonroot
EXPOSE 8604 50144
ENTRYPOINT ["/video-service"]
