package tg

import (
	"fmt"
	"os"
	"path/filepath"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/RandyVentures/tgcli/internal/config"
)

// SendTextOptions for sending text messages.
type SendTextOptions struct {
	ChatID  int64
	Text    string
	ReplyTo int
}

// SendText sends a text message.
func (c *Client) SendText(opts SendTextOptions) (*tgbotapi.Message, error) {
	// Validate message length
	if len(opts.Text) > config.MaxMessageLength {
		return nil, fmt.Errorf("message too long (max %d characters)", config.MaxMessageLength)
	}
	if len(opts.Text) == 0 {
		return nil, fmt.Errorf("message cannot be empty")
	}

	msg := tgbotapi.NewMessage(opts.ChatID, opts.Text)
	if opts.ReplyTo != 0 {
		msg.ReplyToMessageID = opts.ReplyTo
	}

	sent, err := c.bot.Send(msg)
	if err != nil {
		return nil, fmt.Errorf("send message: %w", err)
	}

	return &sent, nil
}

// SendFileOptions for sending files.
type SendFileOptions struct {
	ChatID   int64
	FilePath string
	Caption  string
	ReplyTo  int
}

// validateFilePath validates a file path for security.
func validateFilePath(path string) error {
	// Get absolute path to prevent path traversal
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Resolve symlinks to prevent symlink attacks
	realPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		return fmt.Errorf("cannot resolve path: %w", err)
	}

	// Check file exists and is a regular file
	info, err := os.Stat(realPath)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file")
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("path is not a regular file")
	}

	// Check file size
	if info.Size() > config.MaxFileSize {
		return fmt.Errorf("file too large (max %d MB)", config.MaxFileSize/1024/1024)
	}
	if info.Size() == 0 {
		return fmt.Errorf("file is empty")
	}

	return nil
}

// SendFile sends a file (document).
func (c *Client) SendFile(opts SendFileOptions) (*tgbotapi.Message, error) {
	// Validate file path
	if err := validateFilePath(opts.FilePath); err != nil {
		return nil, err
	}

	// Validate caption length
	if len(opts.Caption) > config.MaxMessageLength {
		return nil, fmt.Errorf("caption too long (max %d characters)", config.MaxMessageLength)
	}

	// Get absolute path
	absPath, _ := filepath.Abs(opts.FilePath)

	doc := tgbotapi.NewDocument(opts.ChatID, tgbotapi.FilePath(absPath))
	if opts.Caption != "" {
		doc.Caption = opts.Caption
	}
	if opts.ReplyTo != 0 {
		doc.ReplyToMessageID = opts.ReplyTo
	}

	sent, err := c.bot.Send(doc)
	if err != nil {
		return nil, fmt.Errorf("send file: %w", err)
	}

	return &sent, nil
}

// SendPhotoOptions for sending photos.
type SendPhotoOptions struct {
	ChatID   int64
	FilePath string
	Caption  string
	ReplyTo  int
}

// SendPhoto sends a photo.
func (c *Client) SendPhoto(opts SendPhotoOptions) (*tgbotapi.Message, error) {
	// Validate file path
	if err := validateFilePath(opts.FilePath); err != nil {
		return nil, err
	}

	// Validate caption length
	if len(opts.Caption) > config.MaxMessageLength {
		return nil, fmt.Errorf("caption too long (max %d characters)", config.MaxMessageLength)
	}

	// Get absolute path
	absPath, _ := filepath.Abs(opts.FilePath)

	photo := tgbotapi.NewPhoto(opts.ChatID, tgbotapi.FilePath(absPath))
	if opts.Caption != "" {
		photo.Caption = opts.Caption
	}
	if opts.ReplyTo != 0 {
		photo.ReplyToMessageID = opts.ReplyTo
	}

	sent, err := c.bot.Send(photo)
	if err != nil {
		return nil, fmt.Errorf("send photo: %w", err)
	}

	return &sent, nil
}

// SetReactionOptions for setting reactions.
type SetReactionOptions struct {
	ChatID    int64
	MessageID int
	Emoji     string
}

// SetReaction sets a reaction on a message.
func (c *Client) SetReaction(opts SetReactionOptions) error {
	// Note: Bot API v5.5 doesn't have native reaction support
	// Reactions require Bot API 7.0+ - would need to use raw API call
	// For now, return not supported error
	return fmt.Errorf("reactions require Bot API 7.0+ (upgrade telegram-bot-api library)")
}
