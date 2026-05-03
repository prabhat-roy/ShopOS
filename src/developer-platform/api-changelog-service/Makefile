.PHONY: build test docker
build:  ; CGO_ENABLED=0 go build -o bin/api-changelog-service .
test:   ; go test ./...
docker: ; docker build -t ghcr.io/shopos/api-changelog-service:dev .
