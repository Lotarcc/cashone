#!/bin/bash
set -e

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Initializing development environment...${NC}"

# Check if .env already exists
if [ -f "$PROJECT_ROOT/.env" ]; then
    echo -e "${YELLOW}Warning: .env file already exists. Skipping...${NC}"
else
    # Copy .env.example to .env
    echo "Creating .env file from example..."
    cp "$PROJECT_ROOT/.env.example" "$PROJECT_ROOT/.env"

    # Generate random JWT secret
    JWT_SECRET=$(openssl rand -base64 32)
    
    # Replace the default JWT secret with the generated one
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        sed -i '' "s/your-secret-key-here/$JWT_SECRET/" "$PROJECT_ROOT/.env"
    else
        # Linux
        sed -i "s/your-secret-key-here/$JWT_SECRET/" "$PROJECT_ROOT/.env"
    fi

    echo -e "${GREEN}Created .env file with secure JWT secret${NC}"
fi

# Create necessary directories
echo "Creating required directories..."
mkdir -p "$PROJECT_ROOT/bin"
mkdir -p "$PROJECT_ROOT/logs"

echo -e "${GREEN}Environment initialization complete!${NC}"
echo "You may now edit .env file to match your local setup"
