.PHONY: build test docker
build:  ; CGO_ENABLED=0 go build -o bin/soc2-evidence-service .
test:   ; go test ./...
docker: ; docker build -t ghcr.io/shopos/soc2-evidence-service:dev .
