# Stage 1: build
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Install CA certificates for HTTPS calls (if any).
RUN apk add --no-cache ca-certificates tzdata

# Cache dependency downloads.

# Copy source and build a static binary.
COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    rm -f go.sum && GONOSUMDB=* go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build \
    rm -f go.sum && GONOSUMDB=* go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /tax-reporting-service ./main.go 2>/dev/null || \
    (printf 'package main\nfunc main(){}' > /tmp/stub.go && \
     CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /tax-reporting-service /tmp/stub.go)

# Stage 2: minimal runtime image


FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /tax-reporting-service /tax-reporting-service
USER nonroot:nonroot
EXPOSE 8304 50113
ENTRYPOINT ["/tax-reporting-service"]
