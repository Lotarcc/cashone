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

# Check if version is provided
if [ -z "$1" ]; then
    echo -e "${RED}Error: Version number is required${NC}"
    echo "Usage: $0 <version>"
    echo "Example: $0 1.0.0"
    exit 1
fi

VERSION=$1

# Validate version format (x.y.z)
if ! [[ $VERSION =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo -e "${RED}Error: Invalid version format${NC}"
    echo "Version must be in format: x.y.z"
    echo "Example: 1.0.0"
    exit 1
fi

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo -e "${RED}Error: Not a git repository${NC}"
    exit 1
fi

# Check if working directory is clean
if ! git diff-index --quiet HEAD --; then
    echo -e "${RED}Error: Working directory is not clean${NC}"
    echo "Please commit or stash your changes first"
    exit 1
fi

echo -e "${YELLOW}Creating release $VERSION...${NC}"

# Update version information
echo -e "${YELLOW}Updating version information...${NC}"
./scripts/build-version.sh

# Create changelog entry
CHANGELOG_FILE="$PROJECT_ROOT/CHANGELOG.md"
if [ ! -f "$CHANGELOG_FILE" ]; then
    echo "# Changelog" > "$CHANGELOG_FILE"
    echo "" >> "$CHANGELOG_FILE"
fi

# Get commit messages since last tag
LAST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
if [ -n "$LAST_TAG" ]; then
    COMMITS=$(git log --pretty=format:"- %s" $LAST_TAG..HEAD)
else
    COMMITS=$(git log --pretty=format:"- %s")
fi

# Create new changelog entry
TEMP_FILE=$(mktemp)
echo "# Changelog" > "$TEMP_FILE"
echo "" >> "$TEMP_FILE"
echo "## [$VERSION] - $(date +%Y-%m-%d)" >> "$TEMP_FILE"
echo "" >> "$TEMP_FILE"
echo "### Added" >> "$TEMP_FILE"
echo "$COMMITS" | grep -i "^- feat" >> "$TEMP_FILE" || true
echo "" >> "$TEMP_FILE"
echo "### Changed" >> "$TEMP_FILE"
echo "$COMMITS" | grep -i "^- change\|^- refactor" >> "$TEMP_FILE" || true
echo "" >> "$TEMP_FILE"
echo "### Fixed" >> "$TEMP_FILE"
echo "$COMMITS" | grep -i "^- fix" >> "$TEMP_FILE" || true
echo "" >> "$TEMP_FILE"
echo "" >> "$TEMP_FILE"
cat "$CHANGELOG_FILE" >> "$TEMP_FILE"
mv "$TEMP_FILE" "$CHANGELOG_FILE"

# Stage changes
git add "$CHANGELOG_FILE"
git add "$PROJECT_ROOT/pkg/version/version.go"

# Create release commit
git commit -m "chore: release version $VERSION"

# Create git tag
git tag -a "v$VERSION" -m "Release $VERSION"

echo -e "${GREEN}Release $VERSION created successfully!${NC}"
echo ""
echo "Next steps:"
echo "1. Review the changes:     git show HEAD"
echo "2. Push the changes:       git push origin main"
echo "3. Push the tag:          git push origin v$VERSION"
echo ""
echo -e "${YELLOW}Don't forget to create a release on GitHub!${NC}"
