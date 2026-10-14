package session

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type SessionMessage struct {
	Role       string `json:"role"`
	Content    string `json:"content,omitempty"`
	ToolName   string `json:"tool_name,omitempty"`
	Action     string `json:"action,omitempty"`
	Details    string `json:"details,omitempty"`
	Status     string `json:"status,omitempty"`
	DiffFile   string `json:"diff_file,omitempty"`
	DiffText   string `json:"diff_text,omitempty"`
	AddCount   int    `json:"add_count,omitempty"`
	DelCount   int    `json:"del_count,omitempty"`
	Duration   string `json:"duration,omitempty"`
	Tokens     int    `json:"tokens,omitempty"`
	Summary    string `json:"summary,omitempty"`
}

type Session struct {
	ID        string           `json:"id"`
	Title     string           `json:"title"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
	Messages  []SessionMessage `json:"messages"`
}

// GetStorageDir returns ~/.svyaz/sessions and ensures it exists
func GetStorageDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".svyaz", "sessions")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

// NewSessionID generates a random 12-char hex session ID
func NewSessionID() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// SaveSession writes the session to disk
func SaveSession(s *Session) error {
	if s == nil || s.ID == "" {
		return fmt.Errorf("invalid session")
	}
	dir, err := GetStorageDir()
	if err != nil {
		return err
	}

	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now()
	}
	s.UpdatedAt = time.Now()

	// If title is empty, derive title from the first user message
	if strings.TrimSpace(s.Title) == "" {
		for _, m := range s.Messages {
			if m.Role == "user" && strings.TrimSpace(m.Content) != "" {
				lines := strings.Split(strings.TrimSpace(m.Content), "\n")
				firstLine := strings.TrimSpace(lines[0])
				if len(firstLine) > 50 {
					firstLine = firstLine[:47] + "..."
				}
				s.Title = firstLine
				break
			}
		}
		if s.Title == "" {
			s.Title = "Session " + s.ID
		}
	}

	filePath := filepath.Join(dir, fmt.Sprintf("%s.json", s.ID))
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

// ListSessions lists all saved sessions sorted by UpdatedAt descending
func ListSessions() ([]Session, error) {
	dir, err := GetStorageDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var sessions []Session
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		filePath := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}
		var s Session
		if err := json.Unmarshal(data, &s); err == nil && s.ID != "" {
			sessions = append(sessions, s)
		}
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].UpdatedAt.After(sessions[j].UpdatedAt)
	})

	return sessions, nil
}

// LoadSession loads a session by ID
func LoadSession(id string) (*Session, error) {
	dir, err := GetStorageDir()
	if err != nil {
		return nil, err
	}
	filePath := filepath.Join(dir, fmt.Sprintf("%s.json", id))
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// DeleteSession removes a session file from disk
func DeleteSession(id string) error {
	dir, err := GetStorageDir()
	if err != nil {
		return err
	}
	filePath := filepath.Join(dir, fmt.Sprintf("%s.json", id))
	return os.Remove(filePath)
}
