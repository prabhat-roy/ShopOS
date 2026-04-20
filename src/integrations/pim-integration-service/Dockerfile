# ---- build stage ------------------------------------------------------------
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates git



COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    rm -f go.sum && GONOSUMDB=* go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build \
    rm -f go.sum && GONOSUMDB=* go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /pim-integration-service ./main.go

# ---- final stage ------------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /pim-integration-service /pim-integration-service
USER nonroot:nonroot
EXPOSE 50195
ENTRYPOINT ["/pim-integration-service"]
