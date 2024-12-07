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

# Get version information
VERSION=${VERSION:-$(git describe --tags --always --dirty || echo "unknown")}
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u '+%Y-%m-%d %H:%M:%S UTC')

# Version file path
VERSION_FILE="$PROJECT_ROOT/app/pkg/version/version.go"

echo -e "${YELLOW}Updating version information...${NC}"
echo "Version:    $VERSION"
echo "Commit:     $GIT_COMMIT"
echo "Build Time: $BUILD_TIME"

# Create version file
cat > "$VERSION_FILE" << EOF
package version

import (
	"fmt"
	"runtime"
)

var (
	// Version is the current version of the application
	Version = "$VERSION"

	// GitCommit is the git commit hash
	GitCommit = "$GIT_COMMIT"

	// BuildTime is the build timestamp
	BuildTime = "$BUILD_TIME"

	// GoVersion is the version of Go used to build the application
	GoVersion = runtime.Version()
)

// Info represents version information
type Info struct {
	Version   string \`json:"version"\`
	GitCommit string \`json:"git_commit"\`
	BuildTime string \`json:"build_time"\`
	GoVersion string \`json:"go_version"\`
	Platform  string \`json:"platform"\`
}

// GetInfo returns version information
func GetInfo() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		BuildTime: BuildTime,
		GoVersion: GoVersion,
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// String returns a string representation of version information
func (i Info) String() string {
	return fmt.Sprintf(
		"Version: %s\nGit Commit: %s\nBuild Time: %s\nGo Version: %s\nPlatform: %s",
		i.Version,
		i.GitCommit,
		i.BuildTime,
		i.GoVersion,
		i.Platform,
	)
}
EOF

echo -e "${GREEN}Version information updated successfully!${NC}"
