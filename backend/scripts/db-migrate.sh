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

# Load environment variables
if [ -f "$PROJECT_ROOT/.env" ]; then
    source "$PROJECT_ROOT/.env"
else
    echo -e "${RED}Error: .env file not found${NC}"
    echo "Please run 'make init' first"
    exit 1
fi

# Function to execute SQL file
execute_sql_file() {
    local file=$1
    echo -e "${YELLOW}Executing $file...${NC}"
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$file"
}

# Function to check if database exists
check_database() {
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -lqt | cut -d \| -f 1 | grep -qw $DB_NAME
}

# Create database if it doesn't exist
create_database() {
    echo -e "${YELLOW}Creating database $DB_NAME...${NC}"
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -c "CREATE DATABASE $DB_NAME;"
}

# Main migration logic
echo -e "${GREEN}Starting database migration...${NC}"

# Check if database exists
if ! check_database; then
    create_database
fi

# Execute initialization script
execute_sql_file "$PROJECT_ROOT/docker/postgres/init.sql"

# Execute any additional migration scripts
if [ -d "$PROJECT_ROOT/migrations" ]; then
    for file in "$PROJECT_ROOT/migrations"/*.sql; do
        if [ -f "$file" ]; then
            execute_sql_file "$file"
        fi
    done
fi

# Execute seed data if in development environment
if [ "$ENV" = "development" ] && [ -d "$PROJECT_ROOT/seeds" ]; then
    echo -e "${YELLOW}Loading seed data...${NC}"
    for file in "$PROJECT_ROOT/seeds"/*.sql; do
        if [ -f "$file" ]; then
            execute_sql_file "$file"
        fi
    done
fi

echo -e "${GREEN}Database migration completed successfully!${NC}"
