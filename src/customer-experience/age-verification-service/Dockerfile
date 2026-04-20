# Stage 1: Build
FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY go.mod ./
COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    rm -f go.sum && GONOSUMDB=* go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build \
    rm -f go.sum && GONOSUMDB=* go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build  -trimpath -ldflags="-s -w" -o /bin/age-verification-service .

# Stage 2: Runtime (scratch — stateless, no deps)


FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /bin/age-verification-service /age-verification-service
USER nonroot:nonroot
EXPOSE 8407
ENTRYPOINT ["/age-verification-service"]
