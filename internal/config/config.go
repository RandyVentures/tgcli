package config

import (
	"os"
	"path/filepath"
)

// DefaultStoreDir returns the default store directory.
func DefaultStoreDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".tgcli"
	}
	return filepath.Join(home, ".tgcli")
}
