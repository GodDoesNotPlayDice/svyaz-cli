package session

import (
	"os"
	"testing"
	"time"
)

func TestSessionSaveListLoadDelete(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "svyaz-sessions-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	sessID := NewSessionID()
	sess := &Session{
		ID:        sessID,
		CreatedAt: time.Now(),
		Messages: []SessionMessage{
			{Role: "user", Content: "Hello world testing session persistence"},
			{Role: "agent", Content: "Response from agent"},
		},
	}

	if err := SaveSession(sess); err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}

	list, err := ListSessions()
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 session, got %d", len(list))
	}
	if list[0].Title != "Hello world testing session persistence" {
		t.Fatalf("unexpected title: %s", list[0].Title)
	}

	loaded, err := LoadSession(sessID)
	if err != nil {
		t.Fatalf("LoadSession failed: %v", err)
	}
	if len(loaded.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(loaded.Messages))
	}

	if err := DeleteSession(sessID); err != nil {
		t.Fatalf("DeleteSession failed: %v", err)
	}

	listAfter, err := ListSessions()
	if err != nil {
		t.Fatalf("ListSessions after delete failed: %v", err)
	}
	if len(listAfter) != 0 {
		t.Fatalf("expected 0 sessions after delete, got %d", len(listAfter))
	}
}
