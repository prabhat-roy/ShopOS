.PHONY: build test docker
build:  ; CGO_ENABLED=0 go build -o bin/seller-tier-service .
test:   ; go test ./...
docker: ; docker build -t ghcr.io/shopos/seller-tier-service:dev .
