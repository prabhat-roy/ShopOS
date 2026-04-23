.PHONY: build test run docker-build docker-run

SERVICE_NAME=storefront-cms-adapter
IMAGE_NAME=enterprise/$(SERVICE_NAME)
VERSION?=latest

build:
	bundle install

test:
	bundle exec rspec

run:
	bundle exec ruby app.rb

docker-build:
	docker build -t $(IMAGE_NAME):$(VERSION) .

docker-run:
	docker run --rm -p 8217:8217 --env-file .env $(IMAGE_NAME):$(VERSION)
