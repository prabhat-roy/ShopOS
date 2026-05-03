.PHONY: build test docker
build:  ; CGO_ENABLED=0 go build -o bin/reorder-service .
test:   ; go test ./...
docker: ; docker build -t ghcr.io/shopos/reorder-service:dev .
