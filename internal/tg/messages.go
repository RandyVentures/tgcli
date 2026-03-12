package tg

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// EditMessageOptions for editing messages.
type EditMessageOptions struct {
	ChatID    int64
	MessageID int
	Text      string
}

// EditMessage edits a text message.
func (c *Client) EditMessage(opts EditMessageOptions) (*tgbotapi.Message, error) {
	if len(opts.Text) > MaxMessageLength {
		return nil, fmt.Errorf("message too long (max %d characters)", MaxMessageLength)
	}
	if len(opts.Text) == 0 {
		return nil, fmt.Errorf("message cannot be empty")
	}

	edit := tgbotapi.NewEditMessageText(opts.ChatID, opts.MessageID, opts.Text)
	
	result, err := c.bot.Send(edit)
	if err != nil {
		return nil, fmt.Errorf("edit message: %w", err)
	}

	return &result, nil
}

// DeleteMessageOptions for deleting messages.
type DeleteMessageOptions struct {
	ChatID    int64
	MessageID int
}

// DeleteMessage deletes a message.
func (c *Client) DeleteMessage(opts DeleteMessageOptions) error {
	del := tgbotapi.NewDeleteMessage(opts.ChatID, opts.MessageID)
	
	_, err := c.bot.Request(del)
	if err != nil {
		return fmt.Errorf("delete message: %w", err)
	}

	return nil
}

// ForwardMessageOptions for forwarding messages.
type ForwardMessageOptions struct {
	ToChatID   int64
	FromChatID int64
	MessageID  int
}

// ForwardMessage forwards a message.
func (c *Client) ForwardMessage(opts ForwardMessageOptions) (*tgbotapi.Message, error) {
	fwd := tgbotapi.NewForward(opts.ToChatID, opts.FromChatID, opts.MessageID)
	
	result, err := c.bot.Send(fwd)
	if err != nil {
		return nil, fmt.Errorf("forward message: %w", err)
	}

	return &result, nil
}

// GetChatOptions for getting chat info.
type GetChatOptions struct {
	ChatID int64
}

// GetChat gets detailed chat information.
func (c *Client) GetChat(opts GetChatOptions) (*tgbotapi.Chat, error) {
	chatConfig := tgbotapi.ChatInfoConfig{
		ChatConfig: tgbotapi.ChatConfig{
			ChatID: opts.ChatID,
		},
	}
	
	chat, err := c.bot.GetChat(chatConfig)
	if err != nil {
		return nil, fmt.Errorf("get chat: %w", err)
	}

	return &chat, nil
}
