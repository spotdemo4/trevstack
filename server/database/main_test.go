package database

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNewRestrictsDatabasePermissions(t *testing.T) {
	// os.UserConfigDir uses XDG_CONFIG_HOME on Linux but HOME on macOS
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp)

	configHome, err := os.UserConfigDir()
	if err != nil {
		t.Fatalf("UserConfigDir() error = %v", err)
	}

	db, err := New(context.Background())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	configDir := filepath.Join(configHome, "trevstack")
	assertMode(t, configDir, 0700)
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
