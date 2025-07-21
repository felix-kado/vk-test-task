.PHONY: help docker-up docker-down test generate swagger lint vet

help:
	@echo "Available commands:"
	@echo "  compose-up     - Start the docker-compose stack"
	@echo "  compose-down   - Stop the docker-compose stack and remove volumes"
	@echo "  test          - Run all tests"
	@echo "  generate      - Generate mocks"
	@echo "  swagger       - Generate Swagger documentation"
	@echo "  lint          - Run golangci-lint"
	@echo "  vet           - Run go vet"

compose-up:
	@echo "Starting Docker containers..."
	@cp -n .env.example .env || true
	docker compose up --build -d

compose-down:
	@echo "Stopping Docker containers..."
	docker compose down -v

test:
	@echo "Running tests..."
	go test -race -covermode=atomic ./...

generate:
	@echo "Generating mocks..."
	go generate ./...

swagger:
	@echo "Generating Swagger documentation..."
	@if ! command -v swag >/dev/null 2>&1 && ! test -f ~/go/bin/swag; then \
		echo "Installing swag..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g cmd/api/main.go -o ./docs; \
	else \
		~/go/bin/swag init -g cmd/api/main.go -o ./docs; \
	fi
	@echo "Swagger documentation generated successfully!"

lint:
	@echo "Running golangci-lint..."
	golangci-lint run ./...

vet:
	@echo "Running go vet..."
	go vet ./...
