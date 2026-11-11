package database

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNewRestrictsDatabasePermissions(t *testing.T) {
	configDir := t.TempDir()

	db, err := New(context.Background(), configDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	assertMode(t, filepath.Join(configDir, "trevstack.db"), 0600)
}

func assertMode(t *testing.T, path string, want os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if got := info.Mode().Perm(); got != want {
		t.Errorf("mode for %s = %04o, want %04o", path, got, want)
	}
}
