// Package config provides configuration constants and defaults for tgcli.
package config

import (
	"os"
	"path/filepath"
	"time"
)

const (
	// DefaultTimeout is the default command timeout for non-sync operations.
	DefaultTimeout = 5 * time.Minute

	// SyncTimeout is the long-poll timeout for sync operations.
	SyncTimeout = 60 // seconds

	// MaxFileSize is the maximum file size for uploads (50MB - Telegram limit).
	MaxFileSize = 50 * 1024 * 1024

	// MaxMessageLength is the maximum message text length (Telegram limit).
	MaxMessageLength = 4096

	// BotTokenEnvVar is the environment variable name for the bot token.
	BotTokenEnvVar = "TGCLI_BOT_TOKEN"
)

// DefaultStoreDir returns the default store directory.
func DefaultStoreDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".tgcli"
	}
	return filepath.Join(home, ".tgcli")
}
