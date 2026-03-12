package app

import (
	"fmt"
)

// Options for creating a new App.
type Options struct {
	StoreDir      string
	Version       string
	JSON          bool
	AllowUnauthed bool
}

// App represents the main application state.
type App struct {
	storeDir      string
	version       string
	json          bool
	allowUnauthed bool
}

// New creates a new App instance.
func New(opts Options) (*App, error) {
	a := &App{
		storeDir:      opts.StoreDir,
		version:       opts.Version,
		json:          opts.JSON,
		allowUnauthed: opts.AllowUnauthed,
	}

	// TODO: Initialize store, telegram client, etc.
	// For now, just validate that the store directory is accessible
	if opts.StoreDir == "" {
		return nil, fmt.Errorf("store directory is required")
	}

	return a, nil
}

// Close cleans up app resources.
func (a *App) Close() {
	// TODO: Close telegram client, database connections, etc.
}

// StoreDir returns the store directory path.
func (a *App) StoreDir() string {
	return a.storeDir
}

// Version returns the app version.
func (a *App) Version() string {
	return a.version
}

// JSON returns whether JSON output is enabled.
func (a *App) JSON() bool {
	return a.json
}
