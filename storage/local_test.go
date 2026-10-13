package storage_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"backpack/storage"
)

func TestLocalStorageOperations(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "backpack_storage_test_*")
	if err != nil {
		t.Fatalf("Failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	ls, err := storage.NewLocalStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed creating LocalStorage: %v", err)
	}

	testFileName := "test_backup.dump"
	testContent := []byte("PostgreSQL backup sample content")

	// 1. Upload
	loc, err := ls.Upload(testFileName, bytes.NewReader(testContent))
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	expectedPath := filepath.Join(tempDir, testFileName)
	if loc != expectedPath {
		t.Errorf("Expected location %s, got %s", expectedPath, loc)
	}

	// 2. Download
	readCloser, err := ls.Download(testFileName)
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}
	defer readCloser.Close()

	downloadedBytes, err := io.ReadAll(readCloser)
	if err != nil {
		t.Fatalf("Reading downloaded stream failed: %v", err)
	}

	if string(downloadedBytes) != string(testContent) {
		t.Errorf("Downloaded content mismatch. Got: %s, Want: %s", string(downloadedBytes), string(testContent))
	}

	// 3. List
	files, err := ls.List("test_")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(files) != 1 || files[0] != testFileName {
		t.Errorf("Unexpected list result: %v", files)
	}

	// 4. Delete
	if err := ls.Delete(testFileName); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	filesAfterDelete, err := ls.List("")
	if err != nil {
		t.Fatalf("List after delete failed: %v", err)
	}
	if len(filesAfterDelete) != 0 {
		t.Errorf("Expected 0 files after delete, got %v", filesAfterDelete)
	}
}
