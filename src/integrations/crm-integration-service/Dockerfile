# Stage 1: Build
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src

COPY go.mod ./

COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    rm -f go.sum && GONOSUMDB=* go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build \
    rm -f go.sum && GONOSUMDB=* go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /crm-integration-service ./main.go

# Stage 2: Final minimal image


FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /crm-integration-service /crm-integration-service
USER nonroot:nonroot
EXPOSE 8903 50172
ENTRYPOINT ["/crm-integration-service"]
