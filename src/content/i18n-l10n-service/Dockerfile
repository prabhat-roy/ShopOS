# ---- Build stage ----
FROM golang:1.24-alpine AS builder
WORKDIR /workspace

# Cache module downloads

# Build binary
COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    rm -f go.sum && GONOSUMDB=* go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build \
    rm -f go.sum && GONOSUMDB=* go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /bin/i18n-l10n-service ./main.go



FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /bin/i18n-l10n-service /i18n-l10n-service
USER nonroot:nonroot
EXPOSE 8606 50146
ENTRYPOINT ["/i18n-l10n-service"]
