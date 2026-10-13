package pipeline

import (
	"compress/gzip"
	"fmt"
	"io"
	"time"

	"backpack/adapters"
	"backpack/storage"
)

// BackupResult holds statistics and metadata for a completed backup.
type BackupResult struct {
	BackupName string
	Location   string
	Duration   time.Duration
}

// BackupPipeline orchestrates the backup process:
// Adapter (dump) -> Compressor (gzip) -> StorageBackend (upload)
type BackupPipeline struct {
	Adapter adapters.DatabaseAdapter
	Storage storage.StorageBackend
}

// NewBackupPipeline creates a new BackupPipeline instance.
func NewBackupPipeline(adapter adapters.DatabaseAdapter, storage storage.StorageBackend) *BackupPipeline {
	return &BackupPipeline{
		Adapter: adapter,
		Storage: storage,
	}
}

// Execute performs the streaming backup:
// streams raw database dump into gzip, which is streamed synchronously to storage without loading entire DB in memory.
func (bp *BackupPipeline) Execute(backupName, backupType string, tables []string) (*BackupResult, error) {
	startTime := time.Now()

	// 1. Initiate dump stream from database adapter
	rawStream, err := bp.Adapter.Backup(backupType, tables)
	if err != nil {
		return nil, fmt.Errorf("database backup initialization failed: %w", err)
	}

	// 2. Set up synchronous in-memory pipe for streaming:
	// rawStream -> gzip.Writer -> pipeWriter === pipeReader -> storage.Upload
	pr, pw := io.Pipe()
	errChan := make(chan error, 1)

	go func() {
		defer rawStream.Close()

		gw := gzip.NewWriter(pw)
		var copyErr error

		_, copyErr = io.Copy(gw, rawStream)
		if closeErr := gw.Close(); closeErr != nil && copyErr == nil {
			copyErr = closeErr
		}

		// Close rawStream to wait for the database dump process to exit and verify its status
		if streamCloseErr := rawStream.Close(); streamCloseErr != nil && copyErr == nil {
			copyErr = streamCloseErr
		}

		if copyErr != nil {
			_ = pw.CloseWithError(copyErr)
			errChan <- copyErr
			return
		}

		_ = pw.Close()
		errChan <- nil
	}()

	// 3. Upload stream through storage backend
	location, err := bp.Storage.Upload(backupName, pr)
	if err != nil {
		_ = pr.CloseWithError(err)
		return nil, fmt.Errorf("storage upload failed: %w", err)
	}

	// 4. Wait for compressor & dump stream goroutine to finish
	if streamErr := <-errChan; streamErr != nil {
		// Clean up partial upload if compression/dump process failed
		_ = bp.Storage.Delete(backupName)
		return nil, fmt.Errorf("streaming compression failed: %w", streamErr)
	}

	duration := time.Since(startTime)
	return &BackupResult{
		BackupName: backupName,
		Location:   location,
		Duration:   duration,
	}, nil
}

// RestorePipeline orchestrates the restore process:
// StorageBackend (download) -> Decompressor (gzip) -> Adapter (restore)
type RestorePipeline struct {
	Adapter adapters.DatabaseAdapter
	Storage storage.StorageBackend
}

// NewRestorePipeline creates a new RestorePipeline instance.
func NewRestorePipeline(adapter adapters.DatabaseAdapter, storage storage.StorageBackend) *RestorePipeline {
	return &RestorePipeline{
		Adapter: adapter,
		Storage: storage,
	}
}

// Execute performs the streaming restore:
// pulls compressed stream from storage, decompresses on the fly, and feeds into database adapter.
func (rp *RestorePipeline) Execute(backupName string, tables []string) error {
	// 1. Download stream from storage
	downloadStream, err := rp.Storage.Download(backupName)
	if err != nil {
		return fmt.Errorf("storage download failed for %s: %w", backupName, err)
	}
	defer downloadStream.Close()

	// 2. Decompress stream on the fly
	gzReader, err := gzip.NewReader(downloadStream)
	if err != nil {
		return fmt.Errorf("failed to initialize gzip decompressor: %w", err)
	}
	defer gzReader.Close()

	// 3. Pipe decompressed stream into database adapter restore
	if err := rp.Adapter.Restore(gzReader, tables); err != nil {
		return fmt.Errorf("database restore failed: %w", err)
	}

	return nil
}
