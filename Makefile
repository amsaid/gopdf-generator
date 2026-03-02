.PHONY: all build run test clean install deps example-server example-cli

# Default target
all: deps build

# Install dependencies
deps:
	go mod download
	go mod tidy

# Build the server
build:
	go build -o bin/gopdf-server cmd/server/main.go

# Run the server
run: build
	./bin/gopdf-server

# Run with custom port
run-port:
	go run cmd/server/main.go -port=3000

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf test-output/
	rm -f *.pdf

# Install the binary
install: build
	cp bin/gopdf-server $(GOPATH)/bin/

# Create necessary directories
setup:
	mkdir -p bin
	mkdir -p fonts
	mkdir -p downloads
	mkdir -p test-output

# Run example server
example-server: build
	@echo "Starting example server on port 8080..."
	./bin/gopdf-server -port=8080

# Generate example PDFs
example-cli: deps
	@echo "Generating example PDFs..."
	go run examples/usage_example.go

# Test API endpoints
test-api:
	@echo "Testing API endpoints..."
	@echo "1. Health check:"
	@curl -s http://localhost:8080/health | jq .
	@echo "\n2. List fonts:"
	@curl -s http://localhost:8080/api/v1/fonts | jq .
	@echo "\n3. Generate PDF:"
	@curl -s -X POST http://localhost:8080/api/v1/generate \
		-H "Content-Type: application/json" \
		-d '{"template":{"page_size":"A4","elements":[{"type":"text","text":"API Test","font":{"size":24}}]}}' | jq .

# Validate all example templates
validate-examples:
	@echo "Validating example templates..."
	@for file in examples/*.json; do \
		echo "Validating $$file..."; \
		curl -s -X POST http://localhost:8080/api/v1/templates/validate \
			-H "Content-Type: application/json" \
			-d @$$file | jq .; \
	done

# Development mode with hot reload (requires air)
dev:
	which air > /dev/null || go install github.com/cosmtrek/air@latest
	air -c .air.toml

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run

# Generate coverage report
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Docker build
docker-build:
	docker build -t gopdf-generator:latest .

# Docker run
docker-run:
	docker run -p 8080:8080 -v $(PWD)/fonts:/app/fonts gopdf-generator:latest

# Help
help:
	@echo "Available targets:"
	@echo "  make deps          - Install Go dependencies"
	@echo "  make build         - Build the server binary"
	@echo "  make run           - Build and run the server"
	@echo "  make run-port      - Run server on custom port"
	@echo "  make test          - Run all tests"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make setup         - Create necessary directories"
	@echo "  make example-server - Run example server"
	@echo "  make example-cli   - Generate example PDFs"
	@echo "  make test-api      - Test API endpoints"
	@echo "  make validate-examples - Validate example templates"
	@echo "  make fmt           - Format Go code"
	@echo "  make lint          - Run linter"
	@echo "  make coverage      - Generate coverage report"
	@echo "  make docker-build  - Build Docker image"
	@echo "  make docker-run    - Run Docker container"
	@echo "  make help          - Show this help"