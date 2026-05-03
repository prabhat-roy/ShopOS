.PHONY: build test docker
build:  ; CGO_ENABLED=0 go build -o bin/chaos-control-service .
test:   ; go test ./...
docker: ; docker build -t ghcr.io/shopos/chaos-control-service:dev .
