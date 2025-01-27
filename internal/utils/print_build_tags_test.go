package utils

import (
	"testing"
)

func TestBuildTagsFormatedString(t *testing.T) {
	// Define the test cases
	tests := []struct {
		buildVersion   string
		buildDate      string
		buildCommit    string
		expectedOutput string
	}{
		{"1.0.0", "2024-11-05", "abcdef1234567890abcdef1234567890abcdef12",
			"Build version: 1.0.0\nBuild date: 2024-11-05\nBuild commit: abcdef1234567890abcdef1234567890abcdef12\n"},
		{"", "", "",
			"Build version: N/A\nBuild date: N/A\nBuild commit: N/A\n"},
		{"1.0.0", "", "",
			"Build version: 1.0.0\nBuild date: N/A\nBuild commit: N/A\n"},
		{"", "2024-11-05", "",
			"Build version: N/A\nBuild date: 2024-11-05\nBuild commit: N/A\n"},
		{"", "", "abcdef1234567890abcdef1234567890abcdef12",
			"Build version: N/A\nBuild date: N/A\nBuild commit: abcdef1234567890abcdef1234567890abcdef12\n"},
	}

	for _, test := range tests {
		t.Run(test.buildVersion+test.buildDate+test.buildCommit, func(t *testing.T) {
			// Call BuildTagsFormatedString and compare result to expected output
			output := BuildTagsFormatedString(test.buildVersion, test.buildDate, test.buildCommit)
			if output != test.expectedOutput {
				t.Errorf("expected:\n%s\nbut got:\n%s", test.expectedOutput, output)
			}
		})
	}
}

func TestGetBuildInfo(t *testing.T) {
	// Define test cases for getBuildInfo
	tests := []struct {
		info     string
		expected string
	}{
		{"1.0.0", "1.0.0"},
		{"", "N/A"},
		{"Some commit", "Some commit"},
	}

	for _, test := range tests {
		t.Run(test.info, func(t *testing.T) {
			result := getBuildInfo(test.info)
			if result != test.expected {
				t.Errorf("expected: %s, got: %s", test.expected, result)
			}
		})
	}
}
