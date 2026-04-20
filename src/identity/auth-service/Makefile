.PHONY: build test run docker-build fmt lint clean

build:
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
	docker build -t shopos/auth-service:latest .
