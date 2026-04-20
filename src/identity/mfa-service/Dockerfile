# ── Stage 1: build ─────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

# Install git for go mod download (some indirect deps need it).
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Cache module downloads separately from source changes.

# Copy all source and compile a statically-linked binary.
COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    rm -f go.sum && GONOSUMDB=* go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build \
    rm -f go.sum && GONOSUMDB=* go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /mfa-service .

# ── Stage 2: runtime ───────────────────────────────────────────────────────────


FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /mfa-service /mfa-service
USER nonroot:nonroot
EXPOSE 8102 50064
ENTRYPOINT ["/mfa-service"]
