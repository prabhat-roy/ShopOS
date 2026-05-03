.PHONY: build test docker
build:  ; CGO_ENABLED=0 go build -o bin/insurance-rider-service .
test:   ; go test ./...
docker: ; docker build -t ghcr.io/shopos/insurance-rider-service:dev .
