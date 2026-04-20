# ── Build stage ────────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

WORKDIR /app


COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    rm -f go.sum && GONOSUMDB=* go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build \
    rm -f go.sum && GONOSUMDB=* go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /b2b-credit-limit-service ./main.go

# ── Runtime stage ───────────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /b2b-credit-limit-service /b2b-credit-limit-service

EXPOSE 8804 50164

USER nonroot:nonroot

ENTRYPOINT ["/b2b-credit-limit-service"]
