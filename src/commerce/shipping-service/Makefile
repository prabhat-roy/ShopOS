SERVICE   := shipping-service
IMAGE     := shopos/$(SERVICE)
VERSION   ?= latest

.PHONY: build test run fmt lint clean docker-build docker-run migrate

build:
	cargo build

build-release:
	cargo build --release

test:
	cargo test

run:
	cargo run

fmt:
	cargo fmt

lint:
	cargo clippy -- -D warnings

clean:
	cargo clean

docker-build:
	docker build -t $(IMAGE):$(VERSION) .

docker-run:
	docker run --rm \
		--env-file .env \
		-p 8138:8138 \
		-p 50084:50084 \
		$(IMAGE):$(VERSION)

migrate:
	sqlx migrate run --database-url "$${DATABASE_URL}"
