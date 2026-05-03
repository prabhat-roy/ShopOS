.PHONY: build test docker
build:  ; CGO_ENABLED=0 go build -o bin/cart-recovery-service .
test:   ; go test ./...
docker: ; docker build -t ghcr.io/shopos/cart-recovery-service:dev .
