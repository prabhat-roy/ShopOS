.PHONY: build test docker
build:  ; CGO_ENABLED=0 go build -o bin/seller-onboarding-service .
test:   ; go test ./...
docker: ; docker build -t ghcr.io/shopos/seller-onboarding-service:dev .
