# ---- build stage ------------------------------------------------------------
FROM golang:1.24-alpine AS builder

# ca-certificates is needed for outbound TLS calls (Redis TLS, etc.)
RUN apk add --no-cache ca-certificates git

WORKDIR /src

# Download dependencies first for layer caching.

# Copy the rest of the source tree.

# Build a statically-linked binary so the runtime image stays minimal.
COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    rm -f go.sum && GONOSUMDB=* go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build \
    rm -f go.sum && GONOSUMDB=* go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /bin/device-fingerprint-service ./main.go

# ---- runtime stage ----------------------------------------------------------


FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /bin/device-fingerprint-service /device-fingerprint-service
USER nonroot:nonroot
EXPOSE 8105 50067
ENTRYPOINT ["/device-fingerprint-service"]
