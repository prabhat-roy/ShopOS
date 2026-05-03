.PHONY: build test docker
build:  ; CGO_ENABLED=0 go build -o bin/catalog-translation-service .
test:   ; go test ./...
docker: ; docker build -t ghcr.io/shopos/catalog-translation-service:dev .
