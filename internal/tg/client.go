package tg

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/session"
	"github.com/RandyVentures/tgcli/internal/store"
)

// Client wraps the Telegram client (gotd).
type Client struct {
	client    *telegram.Client
	api       *tg.Client
	storeDir  string
	storage   *session.FileStorage
	appID     int
	appHash   string
	store     *store.Store
}

// Options for creating a new Client.
type Options struct {
	StoreDir string
	AppID    string
	AppHash  string
	Store    *store.Store
}

// New creates a new Telegram client.
func New(opts Options) (*Client, error) {
	if opts.StoreDir == "" {
		return nil, fmt.Errorf("store directory is required")
	}
	if opts.AppID == "" {
		return nil, fmt.Errorf("TGCLI_APP_ID environment variable is required")
	}
	if opts.AppHash == "" {
		return nil, fmt.Errorf("TGCLI_APP_HASH environment variable is required")
	}
	if opts.Store == nil {
		return nil, fmt.Errorf("store is required")
	}

	appID, err := strconv.Atoi(opts.AppID)
	if err != nil {
		return nil, fmt.Errorf("invalid TGCLI_APP_ID: %w", err)
	}

	// Create session storage
	sessionPath := filepath.Join(opts.StoreDir, "session.json")
	storage := &session.FileStorage{
		Path: sessionPath,
	}

	c := &Client{
		storeDir: opts.StoreDir,
		storage:  storage,
		appID:    appID,
		appHash:  opts.AppHash,
		store:    opts.Store,
	}

	// Create telegram client
	c.client = telegram.NewClient(appID, opts.AppHash, telegram.Options{
		SessionStorage: storage,
	})

	return c, nil
}

// Connect establishes a connection to Telegram.
func (c *Client) Connect(ctx context.Context) error {
	return c.client.Run(ctx, func(ctx context.Context) error {
		c.api = c.client.API()
		return nil
	})
}

// Auth performs authentication (phone + code flow).
func (c *Client) Auth(ctx context.Context, phone string) error {
	return c.client.Run(ctx, func(ctx context.Context) error {
		c.api = c.client.API()

		flow := auth.NewFlow(
			Terminal{},
			auth.SendCodeOptions{},
		)

		if err := c.client.Auth().IfNecessary(ctx, flow); err != nil {
			return fmt.Errorf("auth failed: %w", err)
		}

		return nil
	})
}

// IsAuthed checks if the client is authenticated.
func (c *Client) IsAuthed(ctx context.Context) (bool, error) {
	status, err := c.storage.LoadSession(ctx)
	if err != nil {
		return false, nil // No session file means not authed
	}
	return len(status) > 0, nil
}

// Run executes a function with the authenticated client.
func (c *Client) Run(ctx context.Context, f func(ctx context.Context, api *tg.Client) error) error {
	return c.client.Run(ctx, func(ctx context.Context) error {
		c.api = c.client.API()
		
		// Check if authenticated
		status, err := c.client.Auth().Status(ctx)
		if err != nil {
			return fmt.Errorf("check auth status: %w", err)
		}
		if !status.Authorized {
			return fmt.Errorf("not authenticated. Run 'tgcli auth' first")
		}

		return f(ctx, c.api)
	})
}

// Close stops the client.
func (c *Client) Close() error {
	// Client doesn't need explicit close in gotd
	return nil
}

// Terminal implements auth.UserAuthenticator for terminal-based auth.
type Terminal struct{}

func (Terminal) SignUp(ctx context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, fmt.Errorf("sign up not supported")
}

func (Terminal) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	return nil
}

func (Terminal) Phone(_ context.Context) (string, error) {
	fmt.Print("Enter phone number (with country code): ")
	var phone string
	if _, err := fmt.Scanln(&phone); err != nil {
		return "", err
	}
	return phone, nil
}

func (Terminal) Password(_ context.Context) (string, error) {
	fmt.Print("Enter 2FA password: ")
	var password string
	if _, err := fmt.Scanln(&password); err != nil {
		return "", err
	}
	return password, nil
}

func (Terminal) Code(_ context.Context, _ *tg.AuthSentCode) (string, error) {
	fmt.Print("Enter code: ")
	var code string
	if _, err := fmt.Scanln(&code); err != nil {
		return "", err
	}
	return code, nil
}
