package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"fluxgo/config"
)

func TestLoadReadsDotEnv(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	content := []byte(`
# Application settings
APP_NAME="My App"
export APP_ENV=testing
APP_URL=https://example.test
SERVER_ADDR=:9000
VIEWS_ROOT=web/views
SESSION_COOKIE=my_session
SESSION_LIFETIME_MINUTES=30
SESSION_SECURE=true
DB_HOST=mysql
DB_PORT=3307
DB_DATABASE=app_test
DB_USERNAME=tester
DB_PASSWORD=secret
DB_MAX_IDLE_CONNS=5
DB_MAX_OPEN_CONNS=25
DB_CONN_MAX_LIFETIME_MINUTES=15
DB_RUN_MIGRATIONS=false
`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}

	loaded, err := config.Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if loaded.AppName != "My App" {
		t.Fatalf("expected app name %q, got %q", "My App", loaded.AppName)
	}
	if loaded.AppEnv != "testing" {
		t.Fatalf("expected app environment %q, got %q", "testing", loaded.AppEnv)
	}
	if loaded.AppURL != "https://example.test" {
		t.Fatalf("unexpected app URL %q", loaded.AppURL)
	}
	if loaded.ServerAddr != ":9000" {
		t.Fatalf("expected server address %q, got %q", ":9000", loaded.ServerAddr)
	}
	if loaded.ViewsRoot != "web/views" {
		t.Fatalf("expected views root %q, got %q", "web/views", loaded.ViewsRoot)
	}
	if loaded.SessionCookie != "my_session" {
		t.Fatalf("expected session cookie %q, got %q", "my_session", loaded.SessionCookie)
	}
	if loaded.SessionLifetime != 30*time.Minute || !loaded.SessionSecure {
		t.Fatalf("unexpected session configuration: %+v", loaded)
	}
	if loaded.Database.Host != "mysql" ||
		loaded.Database.Port != "3307" ||
		loaded.Database.Name != "app_test" ||
		loaded.Database.User != "tester" ||
		loaded.Database.Password != "secret" {
		t.Fatalf("unexpected database configuration: %+v", loaded.Database)
	}
	if loaded.Database.MaxIdleConns != 5 ||
		loaded.Database.MaxOpenConns != 25 ||
		loaded.Database.ConnMaxLifetime != 15*time.Minute ||
		loaded.Database.RunMigrations {
		t.Fatalf("unexpected database pool configuration: %+v", loaded.Database)
	}
}

func TestLoadRejectsInvalidSessionConfiguration(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte("SESSION_LIFETIME_MINUTES=never\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := config.Load(path); err == nil {
		t.Fatal("expected invalid session lifetime to fail")
	}
}

func TestLoadRejectsInvalidDatabaseConfiguration(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte("DB_MAX_OPEN_CONNS=zero\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := config.Load(path); err == nil {
		t.Fatal("expected invalid database pool size to fail")
	}
}

func TestEnvironmentOverridesDotEnv(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte("APP_ENV=file\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("APP_ENV", "environment")

	loaded, err := config.Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if loaded.AppEnv != "environment" {
		t.Fatalf("expected OS environment to win, got %q", loaded.AppEnv)
	}
}

func TestLoadUsesDefaultsWhenFileIsMissing(t *testing.T) {
	loaded, err := config.Load(filepath.Join(t.TempDir(), "missing.env"))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if loaded.AppName == "" || loaded.ServerAddr == "" || loaded.ViewsRoot == "" {
		t.Fatalf("expected non-empty defaults, got %+v", loaded)
	}
}
