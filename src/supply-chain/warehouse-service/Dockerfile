# ---- build stage ------------------------------------------------------------
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Download deps first (layer cache friendly)


COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    rm -f go.sum && GONOSUMDB=* go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build \
    rm -f go.sum && GONOSUMDB=* go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /warehouse-service ./main.go

# ---- final stage ------------------------------------------------------------


FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /warehouse-service /warehouse-service
USER nonroot:nonroot
EXPOSE 8202 50102
ENTRYPOINT ["/warehouse-service"]
