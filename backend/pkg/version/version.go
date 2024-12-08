package version

import (
	"fmt"
	"runtime"
)

var (
	// Version is the current version of the application
	Version = "6c15e76-dirty"

	// GitCommit is the git commit hash
	GitCommit = "6c15e76"

	// BuildTime is the build timestamp
	BuildTime = "2024-12-08 20:02:59 UTC"

	// GoVersion is the version of Go used to build the application
	GoVersion = runtime.Version()
)

// Info represents version information
type Info struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
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
