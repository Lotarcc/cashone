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

# Function to install Air
install_air() {
    echo -e "${YELLOW}Installing Air...${NC}"
    if ! go install github.com/air-verse/air@latest; then
        echo -e "${RED}Failed to install Air${NC}"
        exit 1
    fi
    echo -e "${GREEN}Successfully installed Air${NC}"
}

# Function to ensure GOPATH/bin is in PATH
ensure_gopath_bin() {
    if [[ ":$PATH:" != *":$HOME/go/bin:"* ]]; then
        echo -e "${YELLOW}Adding $HOME/go/bin to PATH${NC}"
        export PATH="$PATH:$HOME/go/bin"
    fi
}

# Function to load environment variables with CASHONE_ prefix
load_env() {
    if [ -f "$PROJECT_ROOT/.env" ]; then
        echo -e "${YELLOW}Loading environment variables...${NC}"
        while IFS='=' read -r key value || [ -n "$key" ]; do
            # Skip comments and empty lines
            [[ $key =~ ^#.*$ ]] && continue
            [[ -z "$key" ]] && continue
            
            # Remove leading/trailing whitespace and quotes
            key=$(echo "$key" | xargs)
            value=$(echo "$value" | xargs | sed -e 's/^"//' -e 's/"$//' -e "s/^'//" -e "s/'$//")
            
            # Add CASHONE_ prefix if not already present
            if [[ ! $key =~ ^CASHONE_ ]]; then
                export "CASHONE_$key=$value"
            else
                export "$key=$value"
            fi
        done < "$PROJECT_ROOT/.env"
    else
        echo -e "${RED}No .env file found${NC}"
        exit 1
    fi
}

# Function to check and run migrations if needed
check_migrations() {
    echo -e "${YELLOW}Checking migration status...${NC}"
    
    # Run migration status and capture output
    cd "$PROJECT_ROOT"
    status_output=$(go run cmd/migrate/main.go -command status)
    echo "$status_output"
    
    # Check if there are any pending migrations
    if echo "$status_output" | grep -q "\[ \]"; then
        # Found pending migrations
        read -p "Do you want to run pending migrations? (y/n) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            echo -e "${YELLOW}Running migrations...${NC}"
            go run cmd/migrate/main.go -command up
            if [[ $? -eq 0 ]]; then
                echo -e "${GREEN}Migrations completed successfully${NC}"
            else
                echo -e "${RED}Migration failed${NC}"
                exit 1
            fi
        fi
    else
        echo -e "${GREEN}All migrations are up to date${NC}"
    fi
}

# Main script
echo -e "${YELLOW}Setting up development environment...${NC}"

# Set CONFIG_PATH to point to the env directory
export CONFIG_PATH="$PROJECT_ROOT/config/env"

# Load environment variables
load_env

# Ensure GOPATH/bin is in PATH
ensure_gopath_bin

# Check if Air is installed
if ! command_exists air; then
    install_air
fi

# Create necessary directories
mkdir -p "$PROJECT_ROOT/tmp"

# Start database if not running
if ! docker compose ps | grep -q "db.*running"; then
    echo -e "${YELLOW}Starting database...${NC}"
    docker compose up -d db
    echo -e "${YELLOW}Waiting for database to be ready...${NC}"
    sleep 5
fi

# Check and run migrations if needed
check_migrations

# Update version information
echo -e "${YELLOW}Updating version information...${NC}"
"$PROJECT_ROOT/scripts/build-version.sh"

# Generate API documentation
echo -e "${YELLOW}Generating API documentation...${NC}"
"$PROJECT_ROOT/scripts/generate-docs.sh"

# Start the application with Air
echo -e "${GREEN}Starting application in development mode...${NC}"
echo -e "The application will automatically rebuild when files change."
echo -e "API documentation available at: ${GREEN}http://localhost:8081/swagger/index.html${NC}"

# Change to project root before running Air
cd "$PROJECT_ROOT" && air
