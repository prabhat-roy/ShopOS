.PHONY: build test docker
build:  ; CGO_ENABLED=0 go build -o bin/line-service .
test:   ; go test ./...
docker: ; docker build -t ghcr.io/shopos/line-service:dev .
