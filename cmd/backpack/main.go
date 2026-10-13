package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"backpack/adapters"
	"backpack/config"
	"backpack/logger"
	"backpack/pipeline"
	"backpack/storage"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	cfg := config.LoadFromEnv()

	switch command {
	case "test-connection":
		runTestConnection(cfg, os.Args[2:])
	case "backup":
		runBackup(cfg, os.Args[2:])
	case "restore":
		runRestore(cfg, os.Args[2:])
	case "list":
		runList(cfg, os.Args[2:])
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`BackPack: Database Backup & Restore Utility

Usage:
  backpack <command> [flags]

Available Commands:
  test-connection   Test database connectivity and credentials
  backup            Take a database backup, compress, and store
  restore           Restore a database from a compressed backup
  list              List backups in storage backend
  help              Display this help message

Run 'backpack <command> --help' for details on a specific command.`)
}

func runTestConnection(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("test-connection", flag.ExitOnError)
	host := fs.String("host", cfg.Host, "Database host")
	port := fs.Int("port", cfg.Port, "Database port")
	user := fs.String("user", cfg.User, "Database username")
	password := fs.String("password", cfg.Password, "Database password")
	dbname := fs.String("database", cfg.DBName, "Database name")
	_ = fs.Parse(args)

	logger.InitLogger(cfg.LogFormat, slog.LevelInfo)
	slog.Info("Testing database connection", slog.String("host", *host), slog.Int("port", *port), slog.String("database", *dbname))

	adapter := adapters.NewPostgresAdapter(*host, *port, *user, *password, *dbname)
	if err := adapter.TestConnection(); err != nil {
		slog.Error("Database connection test failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("Database connection successful!")
}

func runBackup(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("backup", flag.ExitOnError)
	host := fs.String("host", cfg.Host, "Database host")
	port := fs.Int("port", cfg.Port, "Database port")
	user := fs.String("user", cfg.User, "Database username")
	password := fs.String("password", cfg.Password, "Database password")
	dbname := fs.String("database", cfg.DBName, "Database name")
	backupType := fs.String("type", "full", "Backup type (full or sql)")
	storagePath := fs.String("path", cfg.StoragePath, "Storage destination path")
	tablesFlag := fs.String("tables", "", "Comma-separated list of tables to back up")
	customName := fs.String("name", "", "Custom backup file name")
	_ = fs.Parse(args)

	logger.InitLogger(cfg.LogFormat, slog.LevelInfo)
	logger.LogBackupStart("postgres", *dbname, *backupType)

	var tables []string
	if *tablesFlag != "" {
		for _, t := range strings.Split(*tablesFlag, ",") {
			if trimmed := strings.TrimSpace(t); trimmed != "" {
				tables = append(tables, trimmed)
			}
		}
	}

	backupName := *customName
	if backupName == "" {
		timestamp := time.Now().Format("2006-01-02_150405")
		backupName = fmt.Sprintf("%s_%s.dump.gz", *dbname, timestamp)
	}

	store, err := storage.NewLocalStorage(*storagePath)
	if err != nil {
		slog.Error("Failed to initialize storage", slog.String("error", err.Error()))
		os.Exit(1)
	}

	adapter := adapters.NewPostgresAdapter(*host, *port, *user, *password, *dbname)
	bp := pipeline.NewBackupPipeline(adapter, store)

	result, err := bp.Execute(backupName, *backupType, tables)
	if err != nil {
		logger.LogBackupFailure(backupName, err, 0)
		os.Exit(1)
	}

	logger.LogBackupSuccess(result.BackupName, result.Location, result.Duration)
}

func runRestore(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("restore", flag.ExitOnError)
	host := fs.String("host", cfg.Host, "Database host")
	port := fs.Int("port", cfg.Port, "Database port")
	user := fs.String("user", cfg.User, "Database username")
	password := fs.String("password", cfg.Password, "Database password")
	dbname := fs.String("database", cfg.DBName, "Database name")
	file := fs.String("file", "", "Backup file to restore (required)")
	storagePath := fs.String("path", cfg.StoragePath, "Storage directory path")
	tablesFlag := fs.String("tables", "", "Comma-separated list of tables to restore")
	_ = fs.Parse(args)

	if *file == "" {
		fmt.Fprintln(os.Stderr, "Error: --file is required for restore.")
		fs.Usage()
		os.Exit(1)
	}

	logger.InitLogger(cfg.LogFormat, slog.LevelInfo)
	slog.Info("Starting restore operation", slog.String("file", *file), slog.String("database", *dbname))

	var tables []string
	if *tablesFlag != "" {
		for _, t := range strings.Split(*tablesFlag, ",") {
			if trimmed := strings.TrimSpace(t); trimmed != "" {
				tables = append(tables, trimmed)
			}
		}
	}

	store, err := storage.NewLocalStorage(*storagePath)
	if err != nil {
		slog.Error("Failed to initialize storage", slog.String("error", err.Error()))
		os.Exit(1)
	}

	adapter := adapters.NewPostgresAdapter(*host, *port, *user, *password, *dbname)
	rp := pipeline.NewRestorePipeline(adapter, store)

	startTime := time.Now()
	if err := rp.Execute(*file, tables); err != nil {
		logger.LogRestoreFailure(*file, err, time.Since(startTime))
		os.Exit(1)
	}

	logger.LogRestoreSuccess(*file, time.Since(startTime))
}

func runList(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	storagePath := fs.String("path", cfg.StoragePath, "Storage directory path")
	prefix := fs.String("prefix", "", "Filename prefix filter")
	_ = fs.Parse(args)

	store, err := storage.NewLocalStorage(*storagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize storage: %v\n", err)
		os.Exit(1)
	}

	files, err := store.List(*prefix)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed listing backups: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("No backups found.")
		return
	}

	fmt.Printf("Backups found in %s:\n", *storagePath)
	for _, f := range files {
		fmt.Printf(" - %s\n", f)
	}
}
