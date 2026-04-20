# ---- build stage ----
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git so go mod download can fetch VCS-based modules.
RUN apk add --no-cache git



COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    rm -f go.sum && GONOSUMDB=* go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build \
    rm -f go.sum && GONOSUMDB=* go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build  -trimpath -ldflags="-s -w" -o /media-asset-service .

# ---- runtime stage ----
# Use alpine (not scratch) so CA certificates are available for MinIO TLS connections.


FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /media-asset-service /media-asset-service
USER nonroot:nonroot
EXPOSE 8600 50140
ENTRYPOINT ["/media-asset-service"]
