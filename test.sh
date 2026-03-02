#!/bin/bash

# Test script for GoPDF Generator

set -e

echo "======================================"
echo "GoPDF Generator Test Script"
echo "======================================"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    echo "Please install Go 1.24 or later: https://golang.org/dl/"
    exit 1
fi

echo "Go version: $(go version)"
echo ""

# Create directories
echo "Creating directories..."
mkdir -p bin fonts downloads test-output

# Download dependencies
echo ""
echo "Downloading dependencies..."
go mod download

# Run tests
echo ""
echo "Running tests..."
go test -v ./pkg/parser/... ./pkg/rtl/... ./pkg/generator/...

# Build server
echo ""
echo "Building server..."
go build -o bin/gopdf-server cmd/server/main.go

echo -e "${GREEN}Build successful!${NC}"
echo ""

# Run example (if examples exist)
if [ -f "examples/usage_example.go" ]; then
    echo "Running usage examples..."
    go run examples/usage_example.go
fi

echo ""
echo "======================================"
echo -e "${GREEN}All tests passed!${NC}"
echo "======================================"
echo ""
echo "To start the server, run:"
echo "  ./bin/gopdf-server"
echo ""
echo "Or with custom options:"
echo "  ./bin/gopdf-server -port=8080 -font-dir=./fonts"
echo ""
echo "API will be available at:"
echo "  http://localhost:8080"
echo ""
