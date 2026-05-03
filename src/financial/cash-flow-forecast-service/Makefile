.PHONY: build test docker
build:  ; CGO_ENABLED=0 go build -o bin/cash-flow-forecast-service .
test:   ; go test ./...
docker: ; docker build -t ghcr.io/shopos/cash-flow-forecast-service:dev .
