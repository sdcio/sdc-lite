package utils

import "os"

func CreateFolder(path string) error {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return err
	}
	return nil
}

func FolderExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false // Folder does not exist
	}
	return err == nil && info.IsDir() // Folder exists and is a directory
}
