package store

import (
	"context"
	"fmt"
	"time"
)

// Chat represents a chat in the database.
type Chat struct {
	ID            int64
	Type          string
	Title         string
	Username      string
	LastMessageID int64
	LastMessageTS int64
	UnreadCount   int
	UpdatedAt     int64
}

// UpsertChat inserts or updates a chat.
func (s *Store) UpsertChat(ctx context.Context, id int64, chatType, title, username string) error {
	now := time.Now().UTC().Unix()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO chats (id, type, title, username, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			type = excluded.type,
			title = excluded.title,
			username = excluded.username,
			updated_at = excluded.updated_at
	`, id, chatType, title, username, now)
	if err != nil {
		return fmt.Errorf("upsert chat: %w", err)
	}
	return nil
}

// GetChat retrieves a chat by ID.
func (s *Store) GetChat(ctx context.Context, id int64) (*Chat, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, type, title, username, last_message_id, last_message_ts, unread_count, updated_at
		FROM chats WHERE id = ?
	`, id)

	var chat Chat
	var lastMsgID, lastMsgTS *int64
	err := row.Scan(&chat.ID, &chat.Type, &chat.Title, &chat.Username, &lastMsgID, &lastMsgTS, &chat.UnreadCount, &chat.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get chat: %w", err)
	}
	if lastMsgID != nil {
		chat.LastMessageID = *lastMsgID
	}
	if lastMsgTS != nil {
		chat.LastMessageTS = *lastMsgTS
	}
	return &chat, nil
}

// ListChats returns all chats ordered by last message.
func (s *Store) ListChats(ctx context.Context, limit int) ([]Chat, error) {
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, type, title, username, last_message_id, last_message_ts, unread_count, updated_at
		FROM chats
		ORDER BY COALESCE(last_message_ts, updated_at) DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("list chats: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close rows: %w", closeErr)
		}
	}()

	var chats []Chat
	for rows.Next() {
		var chat Chat
		var lastMsgID, lastMsgTS *int64
		if err := rows.Scan(&chat.ID, &chat.Type, &chat.Title, &chat.Username, &lastMsgID, &lastMsgTS, &chat.UnreadCount, &chat.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan chat: %w", err)
		}
		if lastMsgID != nil {
			chat.LastMessageID = *lastMsgID
		}
		if lastMsgTS != nil {
			chat.LastMessageTS = *lastMsgTS
		}
		chats = append(chats, chat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	return chats, nil
}

// UpdateChatLastMessage updates the last message info for a chat.
func (s *Store) UpdateChatLastMessage(ctx context.Context, chatID, messageID int64, timestamp int64) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE chats SET last_message_id = ?, last_message_ts = ?, updated_at = ?
		WHERE id = ?
	`, messageID, timestamp, time.Now().UTC().Unix(), chatID)
	if err != nil {
		return fmt.Errorf("update chat last message: %w", err)
	}
	return nil
}
