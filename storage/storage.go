package storage

import "io"

// StorageBackend defines the interface for storing and retrieving backup files.
type StorageBackend interface {
	Upload(name string, stream io.Reader) (string, error)
	Download(name string) (io.ReadCloser, error)
	List(prefix string) ([]string, error)
	Delete(name string) error
}
