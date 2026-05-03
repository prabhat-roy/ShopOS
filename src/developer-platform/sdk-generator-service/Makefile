.PHONY: build test docker
build:  ; CGO_ENABLED=0 go build -o bin/sdk-generator-service .
test:   ; go test ./...
docker: ; docker build -t ghcr.io/shopos/sdk-generator-service:dev .
