.PHONY: build test docker
build:  ; CGO_ENABLED=0 go build -o bin/product-feed-service .
test:   ; go test ./...
docker: ; docker build -t ghcr.io/shopos/product-feed-service:dev .
