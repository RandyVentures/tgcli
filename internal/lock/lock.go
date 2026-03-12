package lock

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// Lock represents a file lock on the store directory.
type Lock struct {
	file *os.File
	path string
}

// Acquire attempts to acquire an exclusive lock on the store directory.
func Acquire(storeDir string) (*Lock, error) {
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		return nil, fmt.Errorf("create store dir: %w", err)
	}

	lockPath := filepath.Join(storeDir, "LOCK")
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("open lock file: %w", err)
	}

	// Try to acquire exclusive lock (non-blocking)
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return nil, fmt.Errorf("another tgcli instance is running (lock held on %s)", lockPath)
	}

	return &Lock{file: f, path: lockPath}, nil
}

// Release releases the lock.
func (l *Lock) Release() error {
	if l.file == nil {
		return nil
	}
	defer func() {
		l.file = nil
	}()

	// Unlock
	if err := syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN); err != nil {
		l.file.Close()
		return err
	}

	return l.file.Close()
}
