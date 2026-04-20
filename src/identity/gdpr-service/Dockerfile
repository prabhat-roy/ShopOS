# ── Stage 1: build ────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

# Install build essentials (git needed for go mod download)
RUN apk add --no-cache git ca-certificates

WORKDIR /src

# Cache dependency downloads separately from source build

# Copy source and build a statically-linked binary
COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    rm -f go.sum && GONOSUMDB=* go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build \
    rm -f go.sum && GONOSUMDB=* go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /bin/gdpr-service .

# ── Stage 2: runtime ──────────────────────────────────────────────────────────


FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /bin/gdpr-service /gdpr-service
USER nonroot:nonroot
EXPOSE 8103 50065
ENTRYPOINT ["/gdpr-service"]
