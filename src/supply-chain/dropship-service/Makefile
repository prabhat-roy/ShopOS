.PHONY: build test docker
build:  ; CGO_ENABLED=0 go build -o bin/dropship-service .
test:   ; go test ./...
docker: ; docker build -t ghcr.io/shopos/dropship-service:dev .
