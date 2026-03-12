package store

import (
	"time"
)

// User represents a Telegram user.
type User struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
	Phone     string `json:"phone,omitempty"`
	IsBot     bool   `json:"is_bot"`
}

// UpsertUser inserts or updates a user.
func (s *Store) UpsertUser(user *User) error {
	now := time.Now().UTC().Unix()
	isBot := 0
	if user.IsBot {
		isBot = 1
	}
	
	_, err := s.db.Exec(`
		INSERT INTO users (id, first_name, last_name, username, phone, is_bot, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			first_name = excluded.first_name,
			last_name = excluded.last_name,
			username = excluded.username,
			phone = excluded.phone,
			is_bot = excluded.is_bot,
			updated_at = excluded.updated_at
	`, user.ID, user.FirstName, user.LastName, user.Username, user.Phone, isBot, now)
	
	return err
}
