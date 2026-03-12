package tg

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"github.com/RandyVentures/tgcli/internal/store"
)

// SyncDialogs fetches all dialogs (chats) and stores them.
func (c *Client) SyncDialogs(ctx context.Context) error {
	return c.Run(ctx, func(ctx context.Context, api *tg.Client) error {
		// Get all dialogs
		dialogs, err := api.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
			OffsetPeer: &tg.InputPeerEmpty{},
			Limit:      100,
		})
		if err != nil {
			return fmt.Errorf("get dialogs: %w", err)
		}

		var dialogSlice *tg.MessagesDialogsSlice
		switch d := dialogs.(type) {
		case *tg.MessagesDialogs:
			// Convert to slice format
			for _, dialog := range d.Dialogs {
				if err := c.processDialog(ctx, dialog, d.Users, d.Chats); err != nil {
					return err
				}
			}
		case *tg.MessagesDialogsSlice:
			dialogSlice = d
			for _, dialog := range dialogSlice.Dialogs {
				if err := c.processDialog(ctx, dialog, dialogSlice.Users, dialogSlice.Chats); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (c *Client) processDialog(ctx context.Context, dialog tg.DialogClass, users []tg.UserClass, chats []tg.ChatClass) error {
	d, ok := dialog.(*tg.Dialog)
	if !ok {
		return nil // Skip non-standard dialogs
	}

	// Extract peer info
	var chatID int64
	var chatType string
	var title string
	var username string

	switch peer := d.Peer.(type) {
	case *tg.PeerUser:
		chatID = peer.UserID
		chatType = "user"
		// Find user in users list
		for _, u := range users {
			if user, ok := u.(*tg.User); ok && user.ID == peer.UserID {
				title = user.FirstName
				if user.LastName != "" {
					title += " " + user.LastName
				}
				if user.Username != "" {
					username = user.Username
				}
				// Store user
				if err := c.store.UpsertUser(&store.User{
					ID:        user.ID,
					FirstName: user.FirstName,
					LastName:  user.LastName,
					Username:  user.Username,
					Phone:     user.Phone,
					IsBot:     user.Bot,
				}); err != nil {
					return fmt.Errorf("store user: %w", err)
				}
				break
			}
		}
	case *tg.PeerChat:
		chatID = peer.ChatID
		chatType = "group"
		// Find chat in chats list
		for _, ch := range chats {
			if chat, ok := ch.(*tg.Chat); ok && chat.ID == peer.ChatID {
				title = chat.Title
				break
			}
		}
	case *tg.PeerChannel:
		chatID = peer.ChannelID
		// Find channel in chats list
		for _, ch := range chats {
			if channel, ok := ch.(*tg.Channel); ok && channel.ID == peer.ChannelID {
				title = channel.Title
				if channel.Username != "" {
					username = channel.Username
				}
				if channel.Broadcast {
					chatType = "channel"
				} else {
					chatType = "supergroup"
				}
				break
			}
		}
	}

	// Store chat
	chat := &store.Chat{
		ID:            chatID,
		Type:          chatType,
		Title:         title,
		Username:      username,
		LastMessageID: d.TopMessage,
		UnreadCount:   d.UnreadCount,
	}

	if err := c.store.UpsertChat(chat); err != nil {
		return fmt.Errorf("store chat: %w", err)
	}

	return nil
}

// SyncChatHistory fetches recent messages for a chat.
func (c *Client) SyncChatHistory(ctx context.Context, chatID int64, limit int) error {
	return c.Run(ctx, func(ctx context.Context, api *tg.Client) error {
		// Determine input peer based on chat type
		chat, err := c.store.GetChat(chatID)
		if err != nil {
			return fmt.Errorf("get chat: %w", err)
		}

		var inputPeer tg.InputPeerClass
		switch chat.Type {
		case "user":
			inputPeer = &tg.InputPeerUser{UserID: chatID}
		case "group":
			inputPeer = &tg.InputPeerChat{ChatID: chatID}
		case "channel", "supergroup":
			inputPeer = &tg.InputPeerChannel{ChannelID: chatID}
		default:
			return fmt.Errorf("unknown chat type: %s", chat.Type)
		}

		// Get message history
		messages, err := api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
			Peer:  inputPeer,
			Limit: limit,
		})
		if err != nil {
			return fmt.Errorf("get history: %w", err)
		}

		var msgs []tg.MessageClass
		switch m := messages.(type) {
		case *tg.MessagesMessages:
			msgs = m.Messages
		case *tg.MessagesMessagesSlice:
			msgs = m.Messages
		case *tg.MessagesChannelMessages:
			msgs = m.Messages
		}

		// Store messages
		for _, msg := range msgs {
			if err := c.processMessage(ctx, msg, chatID); err != nil {
				return err
			}
		}

		return nil
	})
}

func (c *Client) processMessage(ctx context.Context, msgClass tg.MessageClass, chatID int64) error {
	msg, ok := msgClass.(*tg.Message)
	if !ok {
		return nil // Skip service messages
	}

	storeMsg := &store.Message{
		ID:     msg.ID,
		ChatID: chatID,
		Date:   int64(msg.Date),
		Text:   msg.Message,
	}

	if msg.FromID != nil {
		if peerUser, ok := msg.FromID.(*tg.PeerUser); ok {
			storeMsg.FromUserID = peerUser.UserID
		}
	}

	if msg.ReplyTo != nil {
		if replyHeader, ok := msg.ReplyTo.(*tg.MessageReplyHeader); ok {
			storeMsg.ReplyToMessageID = replyHeader.ReplyToMsgID
		}
	}

	if msg.Media != nil {
		// Basic media type detection
		switch msg.Media.(type) {
		case *tg.MessageMediaPhoto:
			storeMsg.MediaType = "photo"
		case *tg.MessageMediaDocument:
			storeMsg.MediaType = "document"
		}
	}

	if err := c.store.InsertMessage(storeMsg); err != nil {
		return fmt.Errorf("store message: %w", err)
	}

	return nil
}
