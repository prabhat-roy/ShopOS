.PHONY: build test docker
build:  ; CGO_ENABLED=0 go build -o bin/secrets-rotation-service .
test:   ; go test ./...
docker: ; docker build -t ghcr.io/shopos/secrets-rotation-service:dev .
