.PHONY: build test docker
build:  ; CGO_ENABLED=0 go build -o bin/tip-service .
test:   ; go test ./...
docker: ; docker build -t ghcr.io/shopos/tip-service:dev .
