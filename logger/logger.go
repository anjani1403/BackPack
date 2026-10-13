package logger

import (
	"log/slog"
	"os"
	"time"
)

// InitLogger initializes the global structured logger.
// format can be "json" or "text".
func InitLogger(format string, level slog.Level) *slog.Logger {
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: level,
	}

	if format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

// LogBackupStart logs the start of a backup operation.
func LogBackupStart(dbType, dbName, backupType string) {
	slog.Info("Backup started",
		slog.String("db_type", dbType),
		slog.String("database", dbName),
		slog.String("backup_type", backupType),
		slog.Time("started_at", time.Now()),
	)
}

// LogBackupSuccess logs completion of a successful backup.
func LogBackupSuccess(backupName, location string, duration time.Duration) {
	slog.Info("Backup completed successfully",
		slog.String("backup_name", backupName),
		slog.String("location", location),
		slog.Duration("duration", duration),
		slog.String("status", "SUCCESS"),
		slog.Time("completed_at", time.Now()),
	)
}

// LogBackupFailure logs a backup failure event.
func LogBackupFailure(backupName string, err error, duration time.Duration) {
	slog.Error("Backup failed",
		slog.String("backup_name", backupName),
		slog.Duration("duration", duration),
		slog.String("status", "FAILED"),
		slog.String("error", err.Error()),
		slog.Time("failed_at", time.Now()),
	)
}

// LogRestoreSuccess logs completion of a successful restore.
func LogRestoreSuccess(backupName string, duration time.Duration) {
	slog.Info("Restore completed successfully",
		slog.String("backup_name", backupName),
		slog.Duration("duration", duration),
		slog.String("status", "SUCCESS"),
		slog.Time("completed_at", time.Now()),
	)
}

// LogRestoreFailure logs a restore failure event.
func LogRestoreFailure(backupName string, err error, duration time.Duration) {
	slog.Error("Restore failed",
		slog.String("backup_name", backupName),
		slog.Duration("duration", duration),
		slog.String("status", "FAILED"),
		slog.String("error", err.Error()),
		slog.Time("failed_at", time.Now()),
	)
}
