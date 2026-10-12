package pipeline

import (
	"backpack/adapters"
)

// BackupPipeline orchestrates the backup process:
// Adapter (dump) -> Compressor (gzip) -> StorageBackend (upload)
type BackupPipeline struct {
	Adapter adapters.DatabaseAdapter
	// Compressor and Storage interfaces will go here
}

func NewBackupPipeline(adapter adapters.DatabaseAdapter) *BackupPipeline {
	return &BackupPipeline{
		Adapter: adapter,
	}
}

func (bp *BackupPipeline) Execute(backupType string, tables []string) error {
	// Phase 1 task: Implement streaming pipeline (io.Pipe, io.Copy, etc.)
	return nil
}

// RestorePipeline orchestrates the restore process:
// StorageBackend (download) -> Decompressor (gzip) -> Adapter (restore)
type RestorePipeline struct {
	Adapter adapters.DatabaseAdapter
	// Decompressor and Storage interfaces will go here
}

func NewRestorePipeline(adapter adapters.DatabaseAdapter) *RestorePipeline {
	return &RestorePipeline{
		Adapter: adapter,
	}
}

func (rp *RestorePipeline) Execute(backupName string, tables []string) error {
	// Phase 1 task: Implement restore pipeline stream
	return nil
}
