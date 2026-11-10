// Package configdir locates the directory where the server keeps its files.
package configdir

import (
	"os"
	"path/filepath"
)

// Get returns trevstack inside the user configuration directory, creating it
// with 0700 permissions, since it holds the database.
func Get() (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(userConfigDir, "trevstack")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	if err := os.Chmod(dir, 0700); err != nil {
		return "", err
	}
	return dir, nil
}
