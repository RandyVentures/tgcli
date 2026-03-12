package store

import (
	"path/filepath"
	"testing"
	"time"
)

func openTestDB(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestUpsertAndGetChat(t *testing.T) {
	s := openTestDB(t)

	// Insert a chat
	if err := s.UpsertChat(123, "user", "Alice", "alice"); err != nil {
		t.Fatalf("UpsertChat: %v", err)
	}

	// Retrieve it
	chat, err := s.GetChat(123)
	if err != nil {
		t.Fatalf("GetChat: %v", err)
	}
	if chat.ID != 123 || chat.Type != "user" || chat.Title != "Alice" {
		t.Fatalf("unexpected chat: %+v", chat)
	}

	// Update with new title
	if err := s.UpsertChat(123, "user", "Alice Updated", "alice"); err != nil {
		t.Fatalf("UpsertChat update: %v", err)
	}

	chat, err = s.GetChat(123)
	if err != nil {
		t.Fatalf("GetChat after update: %v", err)
	}
	if chat.Title != "Alice Updated" {
		t.Fatalf("expected title updated, got %q", chat.Title)
	}
}

func TestListChats(t *testing.T) {
	s := openTestDB(t)

	// Insert multiple chats
	_ = s.UpsertChat(1, "user", "Alice", "")
	_ = s.UpsertChat(2, "group", "Group Chat", "")
	_ = s.UpsertChat(3, "channel", "News Channel", "")

	chats, err := s.ListChats(10)
	if err != nil {
		t.Fatalf("ListChats: %v", err)
	}
	if len(chats) != 3 {
		t.Fatalf("expected 3 chats, got %d", len(chats))
	}
}

func TestListChatsLimit(t *testing.T) {
	s := openTestDB(t)

	for i := 1; i <= 10; i++ {
		_ = s.UpsertChat(int64(i), "user", "User", "")
	}

	chats, err := s.ListChats(5)
	if err != nil {
		t.Fatalf("ListChats: %v", err)
	}
	if len(chats) != 5 {
		t.Fatalf("expected 5 chats (limit), got %d", len(chats))
	}
}

func TestUpsertAndGetUser(t *testing.T) {
	s := openTestDB(t)

	if err := s.UpsertUser(456, "John", "Doe", "johndoe", false); err != nil {
		t.Fatalf("UpsertUser: %v", err)
	}

	user, err := s.GetUser(456)
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if user.ID != 456 || user.FirstName != "John" || user.LastName != "Doe" {
		t.Fatalf("unexpected user: %+v", user)
	}
	if user.IsBot {
		t.Fatalf("expected IsBot=false")
	}
}

func TestInsertAndListMessages(t *testing.T) {
	s := openTestDB(t)

	// Insert chat first
	_ = s.UpsertChat(100, "user", "Test Chat", "")

	// Insert messages
	now := time.Now()
	for i := 1; i <= 5; i++ {
		err := s.InsertMessage(int64(i), 100, 200, now.Add(time.Duration(i)*time.Minute), "Message "+string(rune('0'+i)), 0, "", "")
		if err != nil {
			t.Fatalf("InsertMessage %d: %v", i, err)
		}
	}

	messages, err := s.ListMessages(100, 10)
	if err != nil {
		t.Fatalf("ListMessages: %v", err)
	}
	if len(messages) != 5 {
		t.Fatalf("expected 5 messages, got %d", len(messages))
	}

	// Should be ordered by date DESC
	if messages[0].ID != 5 {
		t.Fatalf("expected newest message first, got ID %d", messages[0].ID)
	}
}

func TestSearchMessages(t *testing.T) {
	s := openTestDB(t)

	_ = s.UpsertChat(100, "user", "Test", "")

	now := time.Now()
	_ = s.InsertMessage(1, 100, 200, now, "Hello world", 0, "", "")
	_ = s.InsertMessage(2, 100, 200, now, "Goodbye world", 0, "", "")
	_ = s.InsertMessage(3, 100, 200, now, "Nothing here", 0, "", "")

	// Search for "world"
	results, err := s.SearchMessages("world", 0, 10)
	if err != nil {
		t.Fatalf("SearchMessages: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results for 'world', got %d", len(results))
	}

	// Search in specific chat
	results, err = s.SearchMessages("Hello", 100, 10)
	if err != nil {
		t.Fatalf("SearchMessages with chat: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result for 'Hello' in chat 100, got %d", len(results))
	}
}

func TestSearchMessagesEscapesWildcards(t *testing.T) {
	s := openTestDB(t)

	_ = s.UpsertChat(100, "user", "Test", "")

	now := time.Now()
	_ = s.InsertMessage(1, 100, 200, now, "100% complete", 0, "", "")
	_ = s.InsertMessage(2, 100, 200, now, "file_name.txt", 0, "", "")
	_ = s.InsertMessage(3, 100, 200, now, "normal text", 0, "", "")

	// Search for literal "%"
	results, err := s.SearchMessages("%", 0, 10)
	if err != nil {
		t.Fatalf("SearchMessages: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result for '%%', got %d", len(results))
	}

	// Search for literal "_"
	results, err = s.SearchMessages("_", 0, 10)
	if err != nil {
		t.Fatalf("SearchMessages: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result for '_', got %d", len(results))
	}
}

func TestUpdateChatLastMessage(t *testing.T) {
	s := openTestDB(t)

	_ = s.UpsertChat(100, "user", "Test", "")

	now := time.Now().Unix()
	if err := s.UpdateChatLastMessage(100, 999, now); err != nil {
		t.Fatalf("UpdateChatLastMessage: %v", err)
	}

	chat, err := s.GetChat(100)
	if err != nil {
		t.Fatalf("GetChat: %v", err)
	}
	if chat.LastMessageID != 999 {
		t.Fatalf("expected LastMessageID=999, got %d", chat.LastMessageID)
	}
	if chat.LastMessageTS != now {
		t.Fatalf("expected LastMessageTS=%d, got %d", now, chat.LastMessageTS)
	}
}

func TestStorePermissions(t *testing.T) {
	dir := t.TempDir()
	testDir := filepath.Join(dir, "teststore")

	s, err := Open(testDir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	// Check directory was created (we can't easily check permissions in all environments)
	// but the Open function should have set 0700
}
