.PHONY: build run dev lint vet fmt docker-build docker-up docker-down clean tidy

## Build the bot binary
build:
	go build -o bin/chisa ./cmd

## Run the bot locally (requires .env)
run: build
	./bin/chisa

## Run with live reload using air (install: go install github.com/air-verse/air@latest)
dev:
	air

## Run go vet
vet:
	go vet ./...

## Run golangci-lint (install: https://golangci-lint.run/usage/install/)
lint:
	golangci-lint run --timeout=5m

## Format all Go source files
fmt:
	gofmt -w .

## Tidy go.mod / go.sum
tidy:
	go mod tidy

## Build Docker image
docker-build:
	docker build -t chisa:local .

## Start local stack (Redis + bot) via docker-compose
docker-up:
	docker compose up --build -d

## Stop local stack
docker-down:
	docker compose down

## Remove built binary
clean:
	rm -rf bin/
