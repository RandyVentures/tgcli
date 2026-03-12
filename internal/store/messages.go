package store

import (
	"time"
)

// Message represents a Telegram message.
type Message struct {
	ID               int    `json:"id"`
	ChatID           int64  `json:"chat_id"`
	FromUserID       int64  `json:"from_user_id,omitempty"`
	Date             int64  `json:"date"`
	Text             string `json:"text,omitempty"`
	ReplyToMessageID int    `json:"reply_to_message_id,omitempty"`
	MediaType        string `json:"media_type,omitempty"`
	MediaPath        string `json:"media_path,omitempty"`
}

// InsertMessage inserts a new message.
func (s *Store) InsertMessage(msg *Message) error {
	now := time.Now().UTC().Unix()
	
	_, err := s.db.Exec(`
		INSERT INTO messages (id, chat_id, from_user_id, date, text, reply_to_message_id, media_type, media_path, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id, chat_id) DO UPDATE SET
			from_user_id = excluded.from_user_id,
			date = excluded.date,
			text = excluded.text,
			reply_to_message_id = excluded.reply_to_message_id,
			media_type = excluded.media_type,
			media_path = excluded.media_path,
			updated_at = excluded.updated_at
	`, msg.ID, msg.ChatID, msg.FromUserID, msg.Date, msg.Text, msg.ReplyToMessageID, msg.MediaType, msg.MediaPath, now)
	
	return err
}

// ListMessages returns messages for a chat, newest first.
func (s *Store) ListMessages(chatID int64, limit int) ([]*Message, error) {
	rows, err := s.db.Query(`
		SELECT id, chat_id, from_user_id, date, text, reply_to_message_id, media_type, media_path
		FROM messages
		WHERE chat_id = ?
		ORDER BY date DESC
		LIMIT ?
	`, chatID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.ChatID, &msg.FromUserID, &msg.Date, &msg.Text, &msg.ReplyToMessageID, &msg.MediaType, &msg.MediaPath); err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}
	
	return messages, rows.Err()
}
