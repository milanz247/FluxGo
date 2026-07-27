package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"fluxgo/config"
)

func TestLoadReadsDotEnv(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	content := []byte(`
# Application settings
APP_NAME="My App"
export APP_ENV=testing
SERVER_ADDR=:9000
VIEWS_ROOT=web/views
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
	if loaded.ServerAddr != ":9000" {
		t.Fatalf("expected server address %q, got %q", ":9000", loaded.ServerAddr)
	}
	if loaded.ViewsRoot != "web/views" {
		t.Fatalf("expected views root %q, got %q", "web/views", loaded.ViewsRoot)
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
