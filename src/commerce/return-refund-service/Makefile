SERVICE   := return-refund-service
IMAGE     := shopos/$(SERVICE)
TAG       ?= latest

.PHONY: build test run docker-build docker-run clean restore

## Restore NuGet packages
restore:
	dotnet restore ReturnRefundService.csproj

## Build the service
build: restore
	dotnet build ReturnRefundService.csproj --configuration Release --no-restore

## Run all unit tests (uses InMemory DB — no Postgres required)
test: restore
	dotnet test ReturnRefundService.csproj \
		--configuration Test \
		--logger "console;verbosity=detailed" \
		--results-directory ./TestResults

## Run the service locally (requires Postgres running — see .env.example)
run:
	ASPNETCORE_ENVIRONMENT=Development dotnet run \
		--project ReturnRefundService.csproj \
		--configuration Debug

## Build the Docker image
docker-build:
	docker build -t $(IMAGE):$(TAG) .

## Run the Docker container locally
docker-run: docker-build
	docker run --rm \
		--name $(SERVICE) \
		-p 8131:8131 \
		--env-file .env.example \
		$(IMAGE):$(TAG)

## Remove build artifacts
clean:
	dotnet clean ReturnRefundService.csproj
	rm -rf bin obj TestResults
