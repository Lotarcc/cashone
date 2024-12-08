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

# Check if description is provided
if [ "$1" = "" ]; then
    echo -e "${RED}Error: Migration description is required${NC}"
    echo "Usage: $0 <description>"
    echo "Example: $0 add_user_settings"
    exit 1
fi

# Convert description to snake case
DESCRIPTION=$(echo "$1" | sed 's/[A-Z]/_\l&/g' | sed 's/^_//' | tr ' ' '_' | tr '[:upper:]' '[:lower:]')

# Get next migration number
NEXT_NUMBER=001
if [ -d "$PROJECT_ROOT/migrations" ]; then
    LAST_MIGRATION=$(ls -1 "$PROJECT_ROOT/migrations/"[0-9]*_*.sql 2>/dev/null | grep -v "_down.sql" | tail -n1 || echo "")
    if [ ! -z "$LAST_MIGRATION" ]; then
        LAST_NUMBER=$(basename "$LAST_MIGRATION" | cut -d'_' -f1)
        NEXT_NUMBER=$(printf "%03d" $((10#$LAST_NUMBER + 1)))
    fi
fi

# Create migration files
UP_MIGRATION="$PROJECT_ROOT/migrations/${NEXT_NUMBER}_${DESCRIPTION}.sql"
DOWN_MIGRATION="$PROJECT_ROOT/migrations/${NEXT_NUMBER}_${DESCRIPTION}_down.sql"

echo -e "${YELLOW}Creating migration files:${NC}"
echo "Up migration:   $UP_MIGRATION"
echo "Down migration: $DOWN_MIGRATION"

# Create up migration
cp "$PROJECT_ROOT/migrations/template_migration.sql" "$UP_MIGRATION"
sed -i "s/XXX/$NEXT_NUMBER/g" "$UP_MIGRATION"
sed -i "s/description/$DESCRIPTION/g" "$UP_MIGRATION"

# Create down migration
cp "$PROJECT_ROOT/migrations/template_migration_down.sql" "$DOWN_MIGRATION"
sed -i "s/XXX/$NEXT_NUMBER/g" "$DOWN_MIGRATION"
sed -i "s/description/$DESCRIPTION/g" "$DOWN_MIGRATION"

echo -e "\n${GREEN}Migration files created successfully!${NC}"
echo -e "\nNext steps:"
echo "1. Edit $UP_MIGRATION to add your migration SQL"
echo "2. Edit $DOWN_MIGRATION to add your rollback SQL"
echo "3. Test your migration:"
echo "   make db-test"
echo "4. Apply your migration:"
echo "   make db-migrate"
echo "5. Verify the changes:"
echo "   make db-status"
echo "   make db-shell"

# Open files in editor if available
if [ ! -z "$EDITOR" ]; then
    echo -e "\n${YELLOW}Opening migration files in your editor...${NC}"
    $EDITOR "$UP_MIGRATION" "$DOWN_MIGRATION"
elif command -v code >/dev/null 2>&1; then
    echo -e "\n${YELLOW}Opening migration files in VS Code...${NC}"
    code "$UP_MIGRATION" "$DOWN_MIGRATION"
else
    echo -e "\n${YELLOW}Please edit the migration files in your preferred editor.${NC}"
fi
