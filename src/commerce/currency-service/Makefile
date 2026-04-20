SERVICE   := currency-service
IMAGE     := shopos/$(SERVICE)
VERSION   ?= latest

.PHONY: install start dev test lint clean docker-build docker-run

install:
	npm install

start:
	npm start

dev:
	npm run dev

test:
	npm test

test-watch:
	npx jest --watch

lint:
	npx eslint src tests index.js --ext .js

clean:
	rm -rf node_modules coverage

docker-build:
	docker build -t $(IMAGE):$(VERSION) .

docker-run:
	docker run --rm \
		--env-file .env \
		-p 8139:8139 \
		$(IMAGE):$(VERSION)
