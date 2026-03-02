#!/bin/bash

# API Test Script for GoPDF Generator

BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "======================================"
echo "GoPDF Generator API Test"
echo "======================================"
echo "Base URL: $BASE_URL"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Check if curl is installed
if ! command -v curl &> /dev/null; then
    echo -e "${RED}Error: curl is not installed${NC}"
    exit 1
fi

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo -e "${YELLOW}Warning: jq is not installed. JSON output will be raw.${NC}"
    JQ_CMD="cat"
else
    JQ_CMD="jq ."
fi

mkdir -p test-output

# Test health endpoint
echo "1. Testing health endpoint..."
curl -s "$BASE_URL/health" | $JQ_CMD
echo ""

# Test fonts endpoint
echo "2. Testing fonts endpoint..."
curl -s "$BASE_URL/api/v1/fonts" | $JQ_CMD
echo ""

# Test template validation
echo "3. Testing template validation..."
curl -s -X POST "$BASE_URL/api/v1/templates/validate" \
  -H "Content-Type: application/json" \
  -d '{"page_size": "A4", "elements": [{"type": "text", "text": "Test"}]}' | $JQ_CMD
echo ""

# Test PDF generation
echo "4. Testing PDF generation..."
curl -s -X POST "$BASE_URL/api/v1/generate/template" \
  -H "Content-Type: application/json" \
  -d '{"page_size": "A4", "margin": {"top": 50, "bottom": 50, "left": 50, "right": 50}, "elements": [{"type": "text", "text": "API Test", "font": {"size": 20}}]}' \
  -o test-output/api-test.pdf

echo -e "${GREEN}PDF saved to: test-output/api-test.pdf${NC}"
echo ""

echo "======================================"
echo -e "${GREEN}API tests completed!${NC}"
echo "======================================"