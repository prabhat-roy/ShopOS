.PHONY: build test docker
build:  ; CGO_ENABLED=0 go build -o bin/distributed-lock-service .
test:   ; go test ./...
docker: ; docker build -t ghcr.io/shopos/distributed-lock-service:dev .
