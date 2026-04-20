# ---- Build stage ----
FROM golang:1.24-alpine AS builder

# Install CA certificates so the binary can make TLS calls (Redis TLS, etc.)
RUN apk add --no-cache ca-certificates git

WORKDIR /build

# Cache dependency downloads separately from source compilation.


# Build a statically-linked binary.
COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    rm -f go.sum && GONOSUMDB=* go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build \
    rm -f go.sum && GONOSUMDB=* go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /session-service ./main.go

# ---- Runtime stage ----


FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /session-service /session-service
USER nonroot:nonroot
EXPOSE 8100 50062
ENTRYPOINT ["/session-service"]
