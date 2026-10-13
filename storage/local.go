package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// LocalStorage implements StorageBackend using the local filesystem.
type LocalStorage struct {
	BasePath string
}

// NewLocalStorage initializes a LocalStorage backend rooted at basePath.
func NewLocalStorage(basePath string) (*LocalStorage, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory %s: %w", basePath, err)
	}
	return &LocalStorage{BasePath: basePath}, nil
}

// Upload writes the input stream to a file in the base storage directory.
func (l *LocalStorage) Upload(name string, stream io.Reader) (string, error) {
	destPath := filepath.Join(l.BasePath, name)
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create directory for %s: %w", destPath, err)
	}

	file, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file %s: %w", destPath, err)
	}
	defer file.Close()

	if _, err := io.Copy(file, stream); err != nil {
		_ = os.Remove(destPath)
		return "", fmt.Errorf("failed writing stream to %s: %w", destPath, err)
	}

	return destPath, nil
}

// Download opens a backup file for reading.
func (l *LocalStorage) Download(name string) (io.ReadCloser, error) {
	filePath := filepath.Join(l.BasePath, name)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed opening file %s: %w", filePath, err)
	}
	return file, nil
}

// List returns all backup filenames matching prefix.
func (l *LocalStorage) List(prefix string) ([]string, error) {
	entries, err := os.ReadDir(l.BasePath)
	if err != nil {
		return nil, fmt.Errorf("failed reading directory %s: %w", l.BasePath, err)
	}

	var results []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) {
			results = append(results, entry.Name())
		}
	}
	return results, nil
}

// Delete deletes the specified backup file.
func (l *LocalStorage) Delete(name string) error {
	filePath := filepath.Join(l.BasePath, name)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed removing file %s: %w", filePath, err)
	}
	return nil
}
