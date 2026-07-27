package cli

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestScanProjectTracksSourceFilesAndIgnoresGeneratedDirectories(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "main.go"), "package main")
	writeTestFile(t, filepath.Join(root, "views", "home.gohtml"), "{{define \"content\"}}home{{end}}")
	writeTestFile(t, filepath.Join(root, ".env"), "APP_ENV=test")
	writeTestFile(t, filepath.Join(root, ".flux", "tmp", "app"), "binary")
	writeTestFile(t, filepath.Join(root, ".git", "HEAD"), "main")
	writeTestFile(t, filepath.Join(root, "README.md"), "docs")

	state, err := scanProject(root)
	if err != nil {
		t.Fatalf("scan project: %v", err)
	}

	for _, path := range []string{"main.go", filepath.Join("views", "home.gohtml"), ".env"} {
		if _, exists := state[path]; !exists {
			t.Fatalf("expected %q to be watched", path)
		}
	}
	if len(state) != 3 {
		t.Fatalf("expected 3 watched files, got %#v", state)
	}
}

func TestStatesEqualDetectsFileChanges(t *testing.T) {
	now := time.Now().UnixNano()
	original := map[string]fileState{"main.go": {modified: now, size: 10}}
	same := map[string]fileState{"main.go": {modified: now, size: 10}}
	changed := map[string]fileState{"main.go": {modified: now + 1, size: 10}}

	if !statesEqual(original, same) {
		t.Fatal("expected identical states to match")
	}
	if statesEqual(original, changed) {
		t.Fatal("expected changed states not to match")
	}
}

func TestWatchedFiles(t *testing.T) {
	for _, name := range []string{"main.go", "home.gohtml", ".env", "go.mod", "go.sum"} {
		if !watchedFile(name) {
			t.Fatalf("expected %q to be watched", name)
		}
	}
	for _, name := range []string{"README.md", "app.log", "binary"} {
		if watchedFile(name) {
			t.Fatalf("did not expect %q to be watched", name)
		}
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
