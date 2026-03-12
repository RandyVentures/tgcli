package tg

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/gotd/td/tg"
)

// SendTextMessage sends a text message to a chat.
func (c *Client) SendTextMessage(ctx context.Context, chatID int64, text string, replyToMsgID int) (int, error) {
	var sentMsgID int

	err := c.Run(ctx, func(ctx context.Context, api *tg.Client) error {
		// Get chat to determine peer type
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

		req := &tg.MessagesSendMessageRequest{
			Peer:     inputPeer,
			Message:  text,
			RandomID: rand.Int63(),
		}

		if replyToMsgID != 0 {
			req.ReplyTo = &tg.InputReplyToMessage{
				ReplyToMsgID: replyToMsgID,
			}
		}

		updates, err := api.MessagesSendMessage(ctx, req)
		if err != nil {
			return fmt.Errorf("send message: %w", err)
		}

		// Extract sent message ID from updates
		switch u := updates.(type) {
		case *tg.Updates:
			for _, update := range u.Updates {
				if msgUpdate, ok := update.(*tg.UpdateMessageID); ok {
					sentMsgID = msgUpdate.ID
					break
				}
			}
		case *tg.UpdateShortSentMessage:
			sentMsgID = u.ID
		}

		return nil
	})

	return sentMsgID, err
}
