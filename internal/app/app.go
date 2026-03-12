package app

import (
	"fmt"
	"os"

	"github.com/RandyVentures/tgcli/internal/store"
	"github.com/RandyVentures/tgcli/internal/tg"
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
	store         *store.Store
	tgClient      *tg.Client
}

// New creates a new App instance.
func New(opts Options) (*App, error) {
	if opts.StoreDir == "" {
		return nil, fmt.Errorf("store directory is required")
	}

	// Open store
	st, err := store.Open(opts.StoreDir)
	if err != nil {
		return nil, fmt.Errorf("open store: %w", err)
	}

	// Create Telegram client
	appID := os.Getenv("TGCLI_APP_ID")
	appHash := os.Getenv("TGCLI_APP_HASH")

	if !opts.AllowUnauthed && (appID == "" || appHash == "") {
		st.Close()
		return nil, fmt.Errorf("TGCLI_APP_ID and TGCLI_APP_HASH environment variables are required")
	}

	var tgClient *tg.Client
	if appID != "" && appHash != "" {
		tgClient, err = tg.New(tg.Options{
			StoreDir: opts.StoreDir,
			AppID:    appID,
			AppHash:  appHash,
			Store:    st,
		})
		if err != nil {
			st.Close()
			return nil, fmt.Errorf("create telegram client: %w", err)
		}
	}

	a := &App{
		storeDir:      opts.StoreDir,
		version:       opts.Version,
		json:          opts.JSON,
		allowUnauthed: opts.AllowUnauthed,
		store:         st,
		tgClient:      tgClient,
	}

	return a, nil
}

// Close cleans up app resources.
func (a *App) Close() {
	if a.tgClient != nil {
		a.tgClient.Close()
	}
	if a.store != nil {
		a.store.Close()
	}
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

// Store returns the store instance.
func (a *App) Store() *store.Store {
	return a.store
}

// TGClient returns the Telegram client instance.
func (a *App) TGClient() *tg.Client {
	return a.tgClient
}
