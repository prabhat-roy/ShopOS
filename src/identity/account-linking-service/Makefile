.PHONY: build test docker
build:  ; CGO_ENABLED=0 go build -o bin/account-linking-service .
test:   ; go test ./...
docker: ; docker build -t ghcr.io/shopos/account-linking-service:dev .
