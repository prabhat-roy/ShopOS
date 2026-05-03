.PHONY: build test docker
build:  ; CGO_ENABLED=0 go build -o bin/zapier-connector-service .
test:   ; go test ./...
docker: ; docker build -t ghcr.io/shopos/zapier-connector-service:dev .
