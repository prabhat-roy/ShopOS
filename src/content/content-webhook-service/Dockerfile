FROM crystallang/crystal:1.13-alpine AS builder

WORKDIR /app

COPY shard.yml shard.lock* ./
RUN shards install --production

COPY src src
RUN crystal build src/main.cr --release --static -o bin/content-webhook-service

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

RUN adduser -D -u 1001 nonroot
USER 1001

WORKDIR /app
COPY --from=builder --chown=nonroot:nonroot /app/bin/content-webhook-service .

EXPOSE 8218

CMD ["./content-webhook-service"]
