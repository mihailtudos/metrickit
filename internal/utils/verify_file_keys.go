// Package utils contains utility functions for various tasks.
// It provides functions for verifying file existence, reading and writing files,
// and other utility functions.
package utils

import "os"

// VerifyFileExists checks if a file exists and is not empty.
func VerifyFileExists(path string) bool {
	f, err := os.Stat(path)
	if err != nil {
		return false
	}

	return f.Size() != 0 && !f.IsDir()
}
