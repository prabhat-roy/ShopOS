# Stage 1: Build
FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    rm -f go.sum && GONOSUMDB=* go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build \
    rm -f go.sum && GONOSUMDB=* go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build  -trimpath -ldflags="-s -w" -o /bin/live-chat-service .

# Stage 2: Runtime (distroless scratch)


FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /bin/live-chat-service /live-chat-service
USER nonroot:nonroot
EXPOSE 8406 50126
ENTRYPOINT ["/live-chat-service"]
