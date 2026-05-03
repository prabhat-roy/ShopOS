.PHONY: build test docker
build:  ; CGO_ENABLED=0 go build -o bin/punchout-service .
test:   ; go test ./...
docker: ; docker build -t ghcr.io/shopos/punchout-service:dev .
