SERVICE   := cart-service
IMAGE     := shopos/$(SERVICE)
TAG       ?= latest

.PHONY: build test run docker-build docker-run clean restore lint

## Restore NuGet packages
restore:
	dotnet restore CartService.csproj

## Build the service
build: restore
	dotnet build CartService.csproj --configuration Release --no-restore

## Run all unit tests
test: restore
	dotnet test CartService.csproj \
		--configuration Test \
		--logger "console;verbosity=detailed" \
		--results-directory ./TestResults

## Run the service locally (reads .env.example via dotenv or manual export)
run:
	ASPNETCORE_ENVIRONMENT=Development dotnet run \
		--project CartService.csproj \
		--configuration Debug

## Build the Docker image
docker-build:
	docker build -t $(IMAGE):$(TAG) .

## Run the Docker container locally
docker-run: docker-build
	docker run --rm \
		--name $(SERVICE) \
		-p 8130:8130 \
		--env-file .env.example \
		$(IMAGE):$(TAG)

## Remove build artifacts
clean:
	dotnet clean CartService.csproj
	rm -rf bin obj TestResults
