#!/bin/bash
set -e

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Create coverage directory if it doesn't exist
mkdir -p "$PROJECT_ROOT/coverage"

echo -e "${YELLOW}Running tests with coverage...${NC}"

# Change to app directory
cd "$PROJECT_ROOT/app"

# Run tests with coverage
go test -coverprofile="$PROJECT_ROOT/coverage/coverage.out" -covermode=atomic ./...

# Generate coverage report in HTML
go tool cover -html="$PROJECT_ROOT/coverage/coverage.out" -o "$PROJECT_ROOT/coverage/coverage.html"

# Calculate coverage percentage
COVERAGE=$(go tool cover -func="$PROJECT_ROOT/coverage/coverage.out" | grep total | awk '{print $3}')

echo -e "${GREEN}Tests completed successfully!${NC}"
echo -e "Coverage: ${YELLOW}$COVERAGE${NC}"
echo -e "Coverage report generated at: ${GREEN}coverage/coverage.html${NC}"

# Check if coverage is below threshold
THRESHOLD=80.0
COVERAGE_NUM=$(echo "$COVERAGE" | sed 's/%//')

if (( $(echo "$COVERAGE_NUM < $THRESHOLD" | bc -l) )); then
    echo -e "${RED}Warning: Coverage is below threshold of ${THRESHOLD}%${NC}"
    exit 1
fi
