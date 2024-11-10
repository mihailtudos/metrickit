package utils

import "os"

func VerifyFileExists(path string) bool {
	f, err := os.Stat(path)
	if err != nil {
		return false
	}

	return f.Size() != 0 && !f.IsDir()
}
