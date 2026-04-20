.PHONY: build test run lint tidy docker-build docker-run

build:
	go build -o bin/api-gateway .

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
	docker build -t shopos/api-gateway:latest .

docker-run:
	docker run --rm -p 8080:8080 --env-file .env shopos/api-gateway:latest
