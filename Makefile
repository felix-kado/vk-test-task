.PHONY: help docker-up docker-down test generate lint vet

help:
	@echo "Available commands:"
	@echo "  docker-up     - Start the docker-compose stack"
	@echo "  docker-down   - Stop the docker-compose stack and remove volumes"
	@echo "  test          - Run all tests"
	@echo "  generate      - Generate mocks"
	@echo "  lint          - Run golangci-lint"
	@echo "  vet           - Run go vet"

docker-up:
	@echo "Starting Docker containers..."
	@cp -n .env.example .env || true
	docker compose up --build -d

docker-down:
	@echo "Stopping Docker containers..."
	docker compose down -v

test:
	@echo "Running tests..."
	go test -race -covermode=atomic ./...

generate:
	@echo "Generating mocks..."
	go generate ./...

lint:
	@echo "Running golangci-lint..."
	golangci-lint run ./...

vet:
	@echo "Running go vet..."
	go vet ./...
