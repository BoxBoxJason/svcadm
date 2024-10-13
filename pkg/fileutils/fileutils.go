package fileutils

import (
	"os"
	"path"

	"github.com/boxboxjason/svcadm/pkg/logger"
)

// WriteToFile writes a string content to a file (creates the file if it does not exist)
func WriteToFile(filename string, content string) error {
	// Create the directory if it does not exist
	dir := path.Dir(filename)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		logger.Debug("Creating", dir)
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(content)
	return err
}

// DeleteFile deletes a file / directory
func DeleteDirectory(path string) error {
	return os.RemoveAll(path)
}

// CreateDirectoryIfNotExists creates a directory if it does not exist
// Returns no error if the directory already exists
func CreateDirectoryIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, os.ModePerm)
	}
	return nil
}

func GetFileContent(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
