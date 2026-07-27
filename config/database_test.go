package config_test

import (
	"testing"
	"time"

	"fluxgo/config"
	MySQLDriver "github.com/go-sql-driver/mysql"
)

func TestDatabaseConfigDSN(t *testing.T) {
	database := config.DatabaseConfig{
		Host:            "db.internal",
		Port:            "3307",
		Name:            "fluxgo",
		User:            "app",
		Password:        "secret",
		Charset:         "utf8mb4",
		MaxIdleConns:    5,
		MaxOpenConns:    20,
		ConnMaxLifetime: time.Hour,
	}

	parsed, err := MySQLDriver.ParseDSN(database.DSN())
	if err != nil {
		t.Fatalf("parse generated DSN: %v", err)
	}
	if parsed.Addr != "db.internal:3307" ||
		parsed.DBName != "fluxgo" ||
		parsed.User != "app" ||
		parsed.Passwd != "secret" ||
		!parsed.ParseTime ||
		parsed.Params["charset"] != "utf8mb4" {
		t.Fatalf("unexpected generated DSN: %+v", parsed)
	}
}
