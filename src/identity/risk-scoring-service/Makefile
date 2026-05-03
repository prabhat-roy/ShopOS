.PHONY: build test docker
build:  ; CGO_ENABLED=0 go build -o bin/risk-scoring-service .
test:   ; go test ./...
docker: ; docker build -t ghcr.io/shopos/risk-scoring-service:dev .
