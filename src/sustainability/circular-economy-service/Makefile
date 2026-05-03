.PHONY: build test docker
build:  ; CGO_ENABLED=0 go build -o bin/circular-economy-service .
test:   ; go test ./...
docker: ; docker build -t ghcr.io/shopos/circular-economy-service:dev .
