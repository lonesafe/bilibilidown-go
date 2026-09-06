package util

import (
	"io"
	"os"
	"path/filepath"
)

// createFile creates a file with its parent directories
func createFile(path string) (io.WriteCloser, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
