.PHONY: build test docker
build:  ; CGO_ENABLED=0 go build -o bin/service-registry-service .
test:   ; go test ./...
docker: ; docker build -t ghcr.io/shopos/service-registry-service:dev .
