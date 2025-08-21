#!/bin/bash
# Test script for Repo Onboarding Copilot

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Running test suite...${NC}"

# Run tests with coverage
echo -e "${YELLOW}Running unit tests...${NC}"
go test -v -race -coverprofile=coverage.out ./...

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed${NC}"
    
    # Generate coverage report
    echo -e "${YELLOW}Generating coverage report...${NC}"
    go tool cover -html=coverage.out -o coverage.html
    echo -e "${GREEN}✓ Coverage report generated: coverage.html${NC}"
    
    # Show coverage summary
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
    echo -e "${GREEN}Total coverage: ${COVERAGE}${NC}"
else
    echo -e "${RED}✗ Tests failed${NC}"
    exit 1
fi