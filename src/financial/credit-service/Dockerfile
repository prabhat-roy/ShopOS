# ---- build stage ----
FROM golang:1.24-alpine AS builder

WORKDIR /src

# Download dependencies first for layer caching.


COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    rm -f go.sum && GONOSUMDB=* go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build \
    rm -f go.sum && GONOSUMDB=* go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -trimpath -o /bin/credit-service .

# ---- final stage ----


FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /bin/credit-service /credit-service
USER nonroot:nonroot
EXPOSE 8306 50115
ENTRYPOINT ["/credit-service"]
