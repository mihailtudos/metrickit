package utils

import "fmt"

func PrintBuildTags(buildVersion, buildDate, buildCommit string) {
	// Output the build information
	fmt.Printf("Build version: %s\n", getBuildInfo(buildVersion))
	fmt.Printf("Build date: %s\n", getBuildInfo(buildDate))
	fmt.Printf("Build commit: %s\n", getBuildInfo(buildCommit))
}

func getBuildInfo(info string) string {
	if info == "" {
		return "N/A"
	}
	return info
}
