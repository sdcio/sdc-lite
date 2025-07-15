package utils

import "os"

func FileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false // Folder does not exist
	}
	return err == nil && !info.IsDir()
}
