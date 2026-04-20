.PHONY: build test run lint tidy docker-build docker-run

build:
	go build -o bin/web-bff .

test:
	go test ./... -v -race -coverprofile=coverage.out

coverage:
	go tool cover -html=coverage.out

run:
	go run .

tidy:
	go mod tidy

lint:
	golangci-lint run ./...

docker-build: tidy
	docker build -t shopos/web-bff:latest .

docker-run:
	docker run --rm -p 8081:8081 --env-file .env shopos/web-bff:latest
