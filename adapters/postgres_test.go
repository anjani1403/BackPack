package adapters_test

import (
	"testing"

	"backpack/adapters"
)

func TestPostgresAdapterInitialization(t *testing.T) {
	adapter := adapters.NewPostgresAdapter("localhost", 5432, "postgres", "secret", "mydb")

	if adapter.Host != "localhost" {
		t.Errorf("expected host localhost, got %s", adapter.Host)
	}
	if adapter.Port != 5432 {
		t.Errorf("expected port 5432, got %d", adapter.Port)
	}
	if adapter.User != "postgres" {
		t.Errorf("expected user postgres, got %s", adapter.User)
	}
	if adapter.Password != "secret" {
		t.Errorf("expected password secret, got %s", adapter.Password)
	}
	if adapter.DBName != "mydb" {
		t.Errorf("expected dbName mydb, got %s", adapter.DBName)
	}
	if adapter.PgDumpPath != "pg_dump" {
		t.Errorf("expected default pg_dump path, got %s", adapter.PgDumpPath)
	}
}
