FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o service .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

RUN adduser -D -u 1001 nonroot
USER 1001

WORKDIR /app
COPY --from=builder --chown=nonroot:nonroot /app/service .

EXPOSE 50303

CMD ["./service"]
