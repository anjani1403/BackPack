package pipeline_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"strings"
	"testing"

	"backpack/pipeline"
)

// mockAdapter simulates DatabaseAdapter
type mockAdapter struct {
	dumpContent    string
	restoredBuffer bytes.Buffer
	backupErr      error
	restoreErr     error
}

func (m *mockAdapter) TestConnection() error {
	return nil
}

func (m *mockAdapter) Backup(backupType string, tables []string) (io.ReadCloser, error) {
	if m.backupErr != nil {
		return nil, m.backupErr
	}
	return io.NopCloser(strings.NewReader(m.dumpContent)), nil
}

func (m *mockAdapter) Restore(stream io.Reader, tables []string) error {
	if m.restoreErr != nil {
		return m.restoreErr
	}
	_, err := io.Copy(&m.restoredBuffer, stream)
	return err
}

func (m *mockAdapter) ListTables() ([]string, error) {
	return []string{"users", "orders"}, nil
}

// mockStorage simulates StorageBackend
type mockStorage struct {
	files map[string][]byte
}

func newMockStorage() *mockStorage {
	return &mockStorage{files: make(map[string][]byte)}
}

func (m *mockStorage) Upload(name string, stream io.Reader) (string, error) {
	data, err := io.ReadAll(stream)
	if err != nil {
		return "", err
	}
	m.files[name] = data
	return "mock://" + name, nil
}

func (m *mockStorage) Download(name string) (io.ReadCloser, error) {
	data, ok := m.files[name]
	if !ok {
		return nil, io.ErrUnexpectedEOF
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (m *mockStorage) List(prefix string) ([]string, error) {
	var keys []string
	for k := range m.files {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	return keys, nil
}

func (m *mockStorage) Delete(name string) error {
	delete(m.files, name)
	return nil
}

func TestBackupAndRestorePipeline(t *testing.T) {
	originalData := "CREATE TABLE users (id SERIAL PRIMARY KEY, name VARCHAR(50));\nINSERT INTO users VALUES (1, 'Alice');"
	adapter := &mockAdapter{dumpContent: originalData}
	storage := newMockStorage()

	bp := pipeline.NewBackupPipeline(adapter, storage)
	backupName := "test_backup.sql.gz"

	// 1. Run Backup
	result, err := bp.Execute(backupName, "custom", nil)
	if err != nil {
		t.Fatalf("BackupPipeline.Execute failed: %v", err)
	}

	if result.Location != "mock://"+backupName {
		t.Errorf("Unexpected location: %s", result.Location)
	}

	// 2. Verify stored data is valid gzip
	storedBytes, ok := storage.files[backupName]
	if !ok || len(storedBytes) == 0 {
		t.Fatalf("Stored file %s is empty or not found", backupName)
	}

	// Check gzip magic header: 0x1f, 0x8b
	if storedBytes[0] != 0x1f || storedBytes[1] != 0x8b {
		t.Fatalf("Stored content is not valid gzip, header: %x %x", storedBytes[0], storedBytes[1])
	}

	// Decompress manually to verify content integrity
	gzReader, err := gzip.NewReader(bytes.NewReader(storedBytes))
	if err != nil {
		t.Fatalf("gzip.NewReader failed: %v", err)
	}
	decompressed, err := io.ReadAll(gzReader)
	_ = gzReader.Close()
	if err != nil {
		t.Fatalf("Failed reading decompressed stream: %v", err)
	}

	if string(decompressed) != originalData {
		t.Fatalf("Decompressed data mismatch.\nGot: %s\nWant: %s", string(decompressed), originalData)
	}

	// 3. Run Restore
	restoreAdapter := &mockAdapter{}
	rp := pipeline.NewRestorePipeline(restoreAdapter, storage)

	if err := rp.Execute(backupName, nil); err != nil {
		t.Fatalf("RestorePipeline.Execute failed: %v", err)
	}

	if restoreAdapter.restoredBuffer.String() != originalData {
		t.Fatalf("Restored content mismatch.\nGot: %s\nWant: %s", restoreAdapter.restoredBuffer.String(), originalData)
	}
}
