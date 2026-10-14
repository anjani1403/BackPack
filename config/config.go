package config

import (
	"os"
	"strconv"
)

// Config represents runtime configuration for BackPack operations.
type Config struct {
	DBType      string
	Host        string
	Port        int
	User        string
	Password    string
	DBName      string
	StorageType string
	StoragePath string
	LogFormat   string
}

// LoadFromEnv loads configuration using environment variables and sensible defaults.
func LoadFromEnv() *Config {
	port := 5432
	if envPort := getEnv("BACKPACK_DB_PORT", os.Getenv("PGPORT")); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}

	return &Config{
		DBType:      getEnv("BACKPACK_DB_TYPE", "postgres"),
		Host:        getEnv("BACKPACK_DB_HOST", getEnv("PGHOST", "localhost")),
		Port:        port,
		User:        getEnv("BACKPACK_DB_USER", getEnv("PGUSER", "admin")),
		Password:    getEnv("BACKPACK_DB_PASSWORD", os.Getenv("PGPASSWORD")),
		DBName:      getEnv("BACKPACK_DB_NAME", getEnv("PGDATABASE", "backpack_test")),
		StorageType: getEnv("BACKPACK_STORAGE_TYPE", "local"),
		StoragePath: getEnv("BACKPACK_STORAGE_PATH", "./backups"),
		LogFormat:   getEnv("BACKPACK_LOG_FORMAT", "text"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
