.PHONY: build test docker
build:  ; CGO_ENABLED=0 go build -o bin/proxy-bid-service .
test:   ; go test ./...
docker: ; docker build -t ghcr.io/shopos/proxy-bid-service:dev .
