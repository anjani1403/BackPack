package adapters_test

import (
	"testing"

	"backpack/adapters"
)

func TestMySQLAdapterInitialization(t *testing.T) {
	adapter := adapters.NewMySQLAdapter("localhost", 3306, "root", "secret", "shop_db")

	if adapter.Host != "localhost" {
		t.Errorf("expected host localhost, got %s", adapter.Host)
	}
	if adapter.Port != 3306 {
		t.Errorf("expected port 3306, got %d", adapter.Port)
	}
	if adapter.User != "root" {
		t.Errorf("expected user root, got %s", adapter.User)
	}
	if adapter.Password != "secret" {
		t.Errorf("expected password secret, got %s", adapter.Password)
	}
	if adapter.DBName != "shop_db" {
		t.Errorf("expected dbName shop_db, got %s", adapter.DBName)
	}
	if adapter.MySQLDumpPath != "mysqldump" {
		t.Errorf("expected default mysqldump path, got %s", adapter.MySQLDumpPath)
	}
}
