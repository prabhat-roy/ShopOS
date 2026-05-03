#!/usr/bin/env bash
# Scaffold a new Go service with the standard ShopOS layout.
# Usage: scripts/new-service.sh <domain> <service-name> <port>
set -euo pipefail
DOMAIN="${1:?domain required}"
NAME="${2:?service name required}"
PORT="${3:?port required}"
ROOT="$(git rev-parse --show-toplevel)"
DIR="${ROOT}/src/${DOMAIN}/${NAME}"
mkdir -p "${DIR}"

cat >"${DIR}/go.mod" <<EOF
module github.com/shopos/${NAME}

go 1.23
EOF

cat >"${DIR}/main.go" <<EOF
package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		w.Write([]byte("# HELP up Service up\n# TYPE up gauge\nup 1\n"))
	})
	srv := &http.Server{Addr: ":${PORT}", Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() { logger.Info("listening", "addr", srv.Addr); _ = srv.ListenAndServe() }()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
EOF

cat >"${DIR}/Dockerfile" <<EOF
FROM golang:1.23-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /app .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /app /app
USER nonroot:nonroot
EXPOSE ${PORT}
ENTRYPOINT ["/app"]
EOF

cat >"${DIR}/Makefile" <<EOF
.PHONY: build test docker
build:  ; CGO_ENABLED=0 go build -o bin/${NAME} .
test:   ; go test ./...
docker: ; docker build -t ghcr.io/shopos/${NAME}:dev .
EOF

cat >"${DIR}/.env.example" <<EOF
PORT=${PORT}
LOG_LEVEL=info
DB_URL=postgresql://${NAME}:secret@postgres-primary.databases.svc:5432/${NAME}
KAFKA_BROKERS=kafka-bootstrap.streaming.svc:9092
EOF

echo "Created ${DIR}"
