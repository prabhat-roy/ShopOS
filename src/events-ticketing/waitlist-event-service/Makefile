.PHONY: build test docker
build:  ; CGO_ENABLED=0 go build -o bin/waitlist-event-service .
test:   ; go test ./...
docker: ; docker build -t ghcr.io/shopos/waitlist-event-service:dev .
