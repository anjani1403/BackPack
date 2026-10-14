package adapters

import "io"

// DatabaseAdapter defines the interface for all database adapters.
type DatabaseAdapter interface {
	// TestConnection tests the connection to the database.
	TestConnection() error

	// Backup performs a backup of the database and returns a byte stream of the raw dump.
	Backup(backupType string, tables []string) (io.ReadCloser, error)

	// Restore restores the database from a byte stream.
	Restore(stream io.Reader, tables []string) error

	// ListTables lists all tables/collections available in the database.
	ListTables() ([]string, error)
}
