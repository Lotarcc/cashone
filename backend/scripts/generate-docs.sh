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

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to install swag
install_swag() {
    echo -e "${YELLOW}Installing swag...${NC}"
    if ! go install github.com/swaggo/swag/cmd/swag@latest; then
        echo -e "${RED}Failed to install swag${NC}"
        exit 1
    fi
    echo -e "${GREEN}Successfully installed swag${NC}"
}

# Function to ensure GOPATH/bin is in PATH
ensure_gopath_bin() {
    if [[ ":$PATH:" != *":$HOME/go/bin:"* ]]; then
        echo -e "${YELLOW}Adding $HOME/go/bin to PATH${NC}"
        export PATH="$PATH:$HOME/go/bin"
    fi
}

# Main script
echo -e "${YELLOW}Checking dependencies...${NC}"

# Ensure GOPATH/bin is in PATH
ensure_gopath_bin

# Check if swag is installed
if ! command_exists swag; then
    install_swag
fi

# Change to app directory
cd "$PROJECT_ROOT/app"

echo -e "${YELLOW}Cleaning existing documentation...${NC}"
rm -rf docs

echo -e "${YELLOW}Generating API documentation...${NC}"
if ! swag init -g cmd/main.go --parseDependency --parseInternal; then
    echo -e "${RED}Failed to generate documentation${NC}"
    exit 1
fi

# Verify docs were generated
if [ ! -f "docs/swagger.json" ] || [ ! -f "docs/swagger.yaml" ] || [ ! -f "docs/docs.go" ]; then
    echo -e "${RED}Documentation files were not generated properly${NC}"
    exit 1
fi

echo -e "${GREEN}API documentation generated successfully!${NC}"
echo -e "Documentation will be available at: ${GREEN}http://localhost:8080/swagger/index.html${NC} when the server is running"

# Create symlink to docs in project root if it doesn't exist
DOCS_ROOT="$PROJECT_ROOT/docs"
if [ ! -d "$DOCS_ROOT" ]; then
    echo -e "${YELLOW}Creating docs symlink in project root...${NC}"
    ln -s "$PROJECT_ROOT/app/docs" "$DOCS_ROOT"
fi

echo -e "\n${GREEN}Documentation setup complete!${NC}"
echo -e "You can now:"
echo -e "1. Run the server:     ${YELLOW}make run${NC}"
echo -e "2. View the docs:      ${YELLOW}make serve-docs${NC}"
echo -e "3. Update the docs:    ${YELLOW}make docs${NC}"
