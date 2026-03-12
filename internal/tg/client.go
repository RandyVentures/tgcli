package tg

import (
	"context"
	"fmt"
)

// Client wraps the Telegram client (gotd).
type Client struct {
	// TODO: Add gotd client instance
}

// Options for creating a new Client.
type Options struct {
	StoreDir string
}

// New creates a new Telegram client.
func New(opts Options) (*Client, error) {
	// TODO: Initialize gotd client with session storage
	if opts.StoreDir == "" {
		return nil, fmt.Errorf("store directory is required")
	}

	c := &Client{
		// TODO: Initialize gotd client
	}

	return c, nil
}

// Close stops the client.
func (c *Client) Close() error {
	// TODO: Stop gotd client
	return nil
}

// Auth performs authentication (phone + code flow).
func (c *Client) Auth(ctx context.Context) error {
	// TODO: Implement phone auth flow with gotd
	return fmt.Errorf("not implemented yet")
}

// IsAuthed checks if the client is authenticated.
func (c *Client) IsAuthed() bool {
	// TODO: Check gotd auth status
	return false
}

// Sync performs a sync operation.
func (c *Client) Sync(ctx context.Context, follow bool) error {
	// TODO: Implement sync logic
	return fmt.Errorf("not implemented yet")
}
