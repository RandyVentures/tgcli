package lock

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAcquireAndRelease(t *testing.T) {
	dir := t.TempDir()

	lk, err := Acquire(dir)
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}

	// Check lock file was created
	lockPath := filepath.Join(dir, "LOCK")
	if _, err := os.Stat(lockPath); os.IsNotExist(err) {
		t.Fatalf("expected LOCK file to exist")
	}

	// Release
	if err := lk.Release(); err != nil {
		t.Fatalf("Release: %v", err)
	}

	// Should be able to acquire again after release
	lk2, err := Acquire(dir)
	if err != nil {
		t.Fatalf("Acquire after release: %v", err)
	}
	_ = lk2.Release()
}

func TestDoubleAcquireFails(t *testing.T) {
	dir := t.TempDir()

	lk1, err := Acquire(dir)
	if err != nil {
		t.Fatalf("Acquire first: %v", err)
	}
	defer lk1.Release()

	// Second acquire should fail
	_, err = Acquire(dir)
	if err == nil {
		t.Fatalf("expected second Acquire to fail")
	}
}

func TestReleaseIdempotent(t *testing.T) {
	dir := t.TempDir()

	lk, err := Acquire(dir)
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}

	// Release multiple times should not error
	if err := lk.Release(); err != nil {
		t.Fatalf("Release 1: %v", err)
	}
	if err := lk.Release(); err != nil {
		t.Fatalf("Release 2: %v", err)
	}
}

func TestAcquireCreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "sub", "nested")

	lk, err := Acquire(subdir)
	if err != nil {
		t.Fatalf("Acquire with nested dir: %v", err)
	}
	defer lk.Release()

	// Check directory was created
	if _, err := os.Stat(subdir); os.IsNotExist(err) {
		t.Fatalf("expected nested directory to be created")
	}
}
