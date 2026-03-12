package tg

import (
	"context"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// SyncOptions for syncing messages.
type SyncOptions struct {
	Follow  bool
	Timeout int
}

// Sync receives updates from Telegram.
func (c *Client) Sync(ctx context.Context, opts SyncOptions) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	if opts.Timeout > 0 {
		u.Timeout = opts.Timeout
	}

	updates := c.bot.GetUpdatesChan(u)

	fmt.Printf("Listening for updates (bot: @%s)...\n", c.bot.Self.UserName)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case update := <-updates:
			if err := c.processUpdate(update); err != nil {
				fmt.Printf("Error processing update: %v\n", err)
			}

			if !opts.Follow {
				// One-shot mode, exit after first batch
				return nil
			}
		}
	}
}

// processUpdate handles a single update.
func (c *Client) processUpdate(update tgbotapi.Update) error {
	if update.Message != nil {
		return c.processMessage(update.Message)
	}
	if update.EditedMessage != nil {
		return c.processMessage(update.EditedMessage)
	}
	if update.ChannelPost != nil {
		return c.processMessage(update.ChannelPost)
	}
	return nil
}

// processMessage handles a message update.
func (c *Client) processMessage(msg *tgbotapi.Message) error {
	ctx := context.Background() // Use background context for storage operations

	// Store the message if we have a store
	if c.store != nil {
		// Store chat
		chatType := "user"
		chatTitle := ""
		if msg.Chat.IsGroup() {
			chatType = "group"
			chatTitle = msg.Chat.Title
		} else if msg.Chat.IsSuperGroup() {
			chatType = "supergroup"
			chatTitle = msg.Chat.Title
		} else if msg.Chat.IsChannel() {
			chatType = "channel"
			chatTitle = msg.Chat.Title
		} else if msg.Chat.IsPrivate() {
			chatType = "user"
			if msg.Chat.FirstName != "" {
				chatTitle = msg.Chat.FirstName
				if msg.Chat.LastName != "" {
					chatTitle += " " + msg.Chat.LastName
				}
			}
		}

		if err := c.store.UpsertChat(ctx, msg.Chat.ID, chatType, chatTitle, msg.Chat.UserName); err != nil {
			return fmt.Errorf("store chat: %w", err)
		}

		// Store user if present
		if msg.From != nil {
			if err := c.store.UpsertUser(ctx, int64(msg.From.ID), msg.From.FirstName, msg.From.LastName, msg.From.UserName, msg.From.IsBot); err != nil {
				return fmt.Errorf("store user: %w", err)
			}
		}

		// Store message
		var fromUserID int64
		if msg.From != nil {
			fromUserID = int64(msg.From.ID)
		}
		var replyToID int
		if msg.ReplyToMessage != nil {
			replyToID = msg.ReplyToMessage.MessageID
		}

		if err := c.store.InsertMessage(
			ctx,
			int64(msg.MessageID),
			msg.Chat.ID,
			fromUserID,
			time.Unix(int64(msg.Date), 0),
			msg.Text,
			replyToID,
			"", // media_type
			"", // media_path
		); err != nil {
			return fmt.Errorf("store message: %w", err)
		}
	}

	// Print to console
	from := "unknown"
	if msg.From != nil {
		from = msg.From.FirstName
		if msg.From.UserName != "" {
			from = "@" + msg.From.UserName
		}
	}

	chatName := msg.Chat.Title
	if chatName == "" && msg.Chat.FirstName != "" {
		chatName = msg.Chat.FirstName
	}

	fmt.Printf("[%s] %s: %s\n", chatName, from, msg.Text)
	return nil
}

// GetUpdates gets recent updates (one-shot).
func (c *Client) GetUpdates(limit int) ([]tgbotapi.Update, error) {
	u := tgbotapi.NewUpdate(0)
	u.Limit = limit
	u.Timeout = 0 // Don't long-poll

	updates, err := c.bot.GetUpdates(u)
	if err != nil {
		return nil, fmt.Errorf("get updates: %w", err)
	}

	return updates, nil
}
