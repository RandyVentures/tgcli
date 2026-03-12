package store

import (
	"database/sql"
	"fmt"
	"time"
)

// Chat represents a Telegram chat (user, group, channel, etc).
type Chat struct {
	ID            int64  `json:"id"`
	Type          string `json:"type"`
	Title         string `json:"title"`
	Username      string `json:"username,omitempty"`
	LastMessageID int    `json:"last_message_id,omitempty"`
	LastMessageTs int64  `json:"last_message_ts,omitempty"`
	UnreadCount   int    `json:"unread_count,omitempty"`
}

// UpsertChat inserts or updates a chat.
func (s *Store) UpsertChat(chat *Chat) error {
	now := time.Now().UTC().Unix()
	
	_, err := s.db.Exec(`
		INSERT INTO chats (id, type, title, username, last_message_id, last_message_ts, unread_count, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			type = excluded.type,
			title = excluded.title,
			username = excluded.username,
			last_message_id = excluded.last_message_id,
			last_message_ts = excluded.last_message_ts,
			unread_count = excluded.unread_count,
			updated_at = excluded.updated_at
	`, chat.ID, chat.Type, chat.Title, chat.Username, chat.LastMessageID, chat.LastMessageTs, chat.UnreadCount, now)
	
	return err
}

// GetChat retrieves a chat by ID.
func (s *Store) GetChat(id int64) (*Chat, error) {
	var chat Chat
	err := s.db.QueryRow(`
		SELECT id, type, title, username, last_message_id, last_message_ts, unread_count
		FROM chats
		WHERE id = ?
	`, id).Scan(&chat.ID, &chat.Type, &chat.Title, &chat.Username, &chat.LastMessageID, &chat.LastMessageTs, &chat.UnreadCount)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("chat not found: %d", id)
	}
	return &chat, err
}

// ListChats returns all chats ordered by last message timestamp.
func (s *Store) ListChats() ([]*Chat, error) {
	rows, err := s.db.Query(`
		SELECT id, type, title, username, last_message_id, last_message_ts, unread_count
		FROM chats
		ORDER BY last_message_ts DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []*Chat
	for rows.Next() {
		var chat Chat
		if err := rows.Scan(&chat.ID, &chat.Type, &chat.Title, &chat.Username, &chat.LastMessageID, &chat.LastMessageTs, &chat.UnreadCount); err != nil {
			return nil, err
		}
		chats = append(chats, &chat)
	}
	
	return chats, rows.Err()
}
