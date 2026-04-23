FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o reports-portal-service .

FROM gcr.io/distroless/static:nonroot
WORKDIR /app
COPY --from=builder /app/reports-portal-service .
EXPOSE 8300
USER nonroot:nonroot
ENTRYPOINT ["/app/reports-portal-service"]
