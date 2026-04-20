# ---- build stage ----
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Cache dependencies

# Copy source and compile a statically linked binary
COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    rm -f go.sum && GONOSUMDB=* go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build \
    rm -f go.sum && GONOSUMDB=* go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build  -ldflags="-s -w" -o brand-service .



FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /build/brand-service /brand-service
USER nonroot:nonroot
EXPOSE 8112 50072
ENTRYPOINT ["/brand-service"]
