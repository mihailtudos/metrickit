/*
Package utils provides utility functions for various common operations.

This package includes functionality to print build information such as
version, date, and commit hash, which can be useful for tracking builds
of the application.
*/
package utils

import "fmt"

// BuildTagsFormatedString outputs the build information to standard output.
//
// It formats the output to show the build version, build date, and build commit
// hash, replacing any empty values with "N/A" for clarity.
//
// Parameters:
//   - buildVersion: A string representing the version of the build.
//   - buildDate: A string representing the date when the build was created.
//   - buildCommit: A string representing the commit hash of the build.
//
// Example usage:
//
//	utils.BuildTagsFormatedString("1.0.0", "2024-11-05", "abcdef1234567890abcdef1234567890abcdef12")
func BuildTagsFormatedString(buildVersion, buildDate, buildCommit string) string {
	// Output the build information
	return fmt.Sprintf("Build version: %s\nBuild date: %s\nBuild commit: %s\n",
		getBuildInfo(buildVersion),
		getBuildInfo(buildDate),
		getBuildInfo(buildCommit))
}

// getBuildInfo returns the provided information string or "N/A" if it is empty.
//
// This helper function is used to standardize the output of build information
// by ensuring that no empty strings are printed in the output.
//
// Parameters:
//   - info: A string containing the information to check.
//
// Returns:
//   - A string containing the original info or "N/A" if info is empty.
//
// Example usage:
//
//	info := getBuildInfo("") // returns "N/A"
//	info := getBuildInfo("1.0.0") // returns "1.0.0"
func getBuildInfo(info string) string {
	if info == "" {
		return "N/A"
	}
	return info
}
