package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// Message represents a message in the database.
type Message struct {
	ID               int64
	ChatID           int64
	FromUserID       int64
	Date             int64
	Text             string
	ReplyToMessageID int
	MediaType        string
	MediaPath        string
	UpdatedAt        int64
	Snippet          string // For FTS search results with highlighting
}

// InsertMessage inserts a new message.
func (s *Store) InsertMessage(ctx context.Context, id, chatID, fromUserID int64, date time.Time, text string, replyToID int, mediaType, mediaPath string) error {
	// Truncate text if too long (Telegram limit is 4096)
	if len(text) > MaxMessageLength {
		text = text[:MaxMessageLength]
	}

	now := time.Now().UTC().Unix()
	dateUnix := date.Unix()

	_, err := s.db.ExecContext(ctx, `
		INSERT OR REPLACE INTO messages (id, chat_id, from_user_id, date, text, reply_to_message_id, media_type, media_path, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, chatID, fromUserID, dateUnix, text, replyToID, mediaType, mediaPath, now)
	if err != nil {
		return fmt.Errorf("insert message: %w", err)
	}

	// Update chat's last message
	_ = s.UpdateChatLastMessage(ctx, chatID, id, dateUnix)

	return nil
}

// GetMessage retrieves a message by ID.
func (s *Store) GetMessage(ctx context.Context, id int64) (*Message, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, chat_id, from_user_id, date, text, reply_to_message_id, media_type, media_path, updated_at
		FROM messages WHERE id = ?
	`, id)

	var msg Message
	var text, mediaType, mediaPath sql.NullString
	err := row.Scan(&msg.ID, &msg.ChatID, &msg.FromUserID, &msg.Date, &text, &msg.ReplyToMessageID, &mediaType, &mediaPath, &msg.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get message: %w", err)
	}
	msg.Text = text.String
	msg.MediaType = mediaType.String
	msg.MediaPath = mediaPath.String
	return &msg, nil
}

// ListMessagesParams contains parameters for listing messages.
type ListMessagesParams struct {
	ChatID    int64
	Limit     int
	Before    *time.Time // Messages before this time
	After     *time.Time // Messages after this time
	MediaType string     // Filter by media type
}

// ListMessages returns messages for a chat with optional filters.
func (s *Store) ListMessages(ctx context.Context, params ListMessagesParams) ([]Message, error) {
	limit := params.Limit
	if limit <= 0 {
		limit = DefaultLimit
	}
	// Cap limit to prevent abuse
	if limit > MaxLimit {
		limit = MaxLimit
	}

	query := `
		SELECT id, chat_id, from_user_id, date, text, reply_to_message_id, media_type, media_path, updated_at
		FROM messages
		WHERE chat_id = ?`
	
	args := []interface{}{params.ChatID}

	if params.Before != nil {
		query += " AND date < ?"
		args = append(args, params.Before.Unix())
	}
	if params.After != nil {
		query += " AND date > ?"
		args = append(args, params.After.Unix())
	}
	if params.MediaType != "" {
		query += " AND media_type = ?"
		args = append(args, params.MediaType)
	}

	query += " ORDER BY date DESC LIMIT ?"
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close rows: %w", closeErr)
		}
	}()

	var messages []Message
	for rows.Next() {
		var msg Message
		var text, mediaType, mediaPath sql.NullString
		if err := rows.Scan(&msg.ID, &msg.ChatID, &msg.FromUserID, &msg.Date, &text, &msg.ReplyToMessageID, &mediaType, &mediaPath, &msg.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		msg.Text = text.String
		msg.MediaType = mediaType.String
		msg.MediaPath = mediaPath.String
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	return messages, nil
}

// escapeLikePattern escapes special characters in LIKE patterns to prevent injection.
func escapeLikePattern(s string) string {
	// Escape special LIKE characters: % _ \
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "%", "\\%")
	s = strings.ReplaceAll(s, "_", "\\_")
	return s
}

// SearchMessagesParams contains parameters for searching messages.
type SearchMessagesParams struct {
	Query     string
	ChatID    int64
	Limit     int
	Before    *time.Time // Messages before this time
	After     *time.Time // Messages after this time
	MediaType string     // Filter by media type
}

// SearchMessages searches messages using FTS if available, otherwise LIKE.
func (s *Store) SearchMessages(ctx context.Context, params SearchMessagesParams) ([]Message, error) {
	if strings.TrimSpace(params.Query) == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	limit := params.Limit
	if limit <= 0 {
		limit = DefaultLimit
	}
	// Cap limit to prevent abuse
	if limit > MaxLimit {
		limit = MaxLimit
	}

	// Use FTS if available, otherwise fall back to LIKE
	if s.ftsEnabled {
		return s.searchMessagesFTS(ctx, params, limit)
	}
	return s.searchMessagesLIKE(ctx, params, limit)
}

// searchMessagesFTS performs full-text search using FTS5.
func (s *Store) searchMessagesFTS(ctx context.Context, params SearchMessagesParams, limit int) ([]Message, error) {
	query := `
		SELECT m.id, m.chat_id, m.from_user_id, m.date, m.text, m.reply_to_message_id, 
		       m.media_type, m.media_path, m.updated_at,
		       snippet(messages_fts, 0, '[', ']', '…', 12) as snippet
		FROM messages_fts
		JOIN messages m ON messages_fts.rowid = m.rowid
		WHERE messages_fts MATCH ?`
	
	args := []interface{}{params.Query}

	if params.ChatID != 0 {
		query += " AND m.chat_id = ?"
		args = append(args, params.ChatID)
	}
	if params.Before != nil {
		query += " AND m.date < ?"
		args = append(args, params.Before.Unix())
	}
	if params.After != nil {
		query += " AND m.date > ?"
		args = append(args, params.After.Unix())
	}
	if params.MediaType != "" {
		query += " AND m.media_type = ?"
		args = append(args, params.MediaType)
	}

	query += " ORDER BY bm25(messages_fts) LIMIT ?"
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("FTS search: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close rows: %w", closeErr)
		}
	}()

	var messages []Message
	for rows.Next() {
		var msg Message
		var text, mediaType, mediaPath, snippet sql.NullString
		if err := rows.Scan(&msg.ID, &msg.ChatID, &msg.FromUserID, &msg.Date, &text, &msg.ReplyToMessageID, &mediaType, &mediaPath, &msg.UpdatedAt, &snippet); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		msg.Text = text.String
		msg.MediaType = mediaType.String
		msg.MediaPath = mediaPath.String
		msg.Snippet = snippet.String
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	return messages, nil
}

// searchMessagesLIKE performs search using LIKE (fallback when FTS unavailable).
func (s *Store) searchMessagesLIKE(ctx context.Context, params SearchMessagesParams, limit int) ([]Message, error) {
	// Escape LIKE wildcards in user input to prevent LIKE injection
	escapedQuery := escapeLikePattern(params.Query)
	searchPattern := "%" + escapedQuery + "%"

	query := `
		SELECT id, chat_id, from_user_id, date, text, reply_to_message_id, media_type, media_path, updated_at
		FROM messages
		WHERE text LIKE ? ESCAPE '\'`
	
	args := []interface{}{searchPattern}

	if params.ChatID != 0 {
		query += " AND chat_id = ?"
		args = append(args, params.ChatID)
	}
	if params.Before != nil {
		query += " AND date < ?"
		args = append(args, params.Before.Unix())
	}
	if params.After != nil {
		query += " AND date > ?"
		args = append(args, params.After.Unix())
	}
	if params.MediaType != "" {
		query += " AND media_type = ?"
		args = append(args, params.MediaType)
	}

	query += " ORDER BY date DESC LIMIT ?"
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("LIKE search: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close rows: %w", closeErr)
		}
	}()

	var messages []Message
	for rows.Next() {
		var msg Message
		var text, mediaType, mediaPath sql.NullString
		if err := rows.Scan(&msg.ID, &msg.ChatID, &msg.FromUserID, &msg.Date, &text, &msg.ReplyToMessageID, &mediaType, &mediaPath, &msg.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		msg.Text = text.String
		msg.MediaType = mediaType.String
		msg.MediaPath = mediaPath.String
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	return messages, nil
}
