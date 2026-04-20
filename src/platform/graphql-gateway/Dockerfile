# ── Stage 1: build ────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

# Install ca-certificates for TLS calls to downstream services.
RUN apk add --no-cache ca-certificates git

WORKDIR /src

# Cache dependency download separately from source compilation.
COPY go.mod ./

# Copy source and build a fully static binary.
COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    rm -f go.sum && GONOSUMDB=* go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build \
    rm -f go.sum && GONOSUMDB=* go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/graphql-gateway .

# ── Stage 2: runtime ──────────────────────────────────────────────────────────


FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/graphql-gateway /graphql-gateway
USER nonroot:nonroot
EXPOSE 8086
ENTRYPOINT ["/graphql-gateway"]
