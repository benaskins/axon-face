package face

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	talk "github.com/benaskins/axon-loop" // Message aliased from axon-talk
)

// Session represents a persisted conversation session.
// The State field is an escape hatch for app-specific data —
// apps marshal their own types into it (sections, approvals, etc.).
type Session struct {
	ID        string            `json:"id"`
	Messages  []talk.Message    `json:"messages"`
	Phase     string            `json:"phase,omitempty"`
	State     map[string]any    `json:"state,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Complete  bool              `json:"complete"`
}

// NewSession creates a session with a timestamped ID.
func NewSession() *Session {
	now := time.Now()
	return &Session{
		ID:        now.Format("2006-01-02T15-04-05"),
		Phase:     "interview",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Save writes the session to dir as a JSON file.
func (s *Session) Save(dir string) error {
	s.UpdatedAt = time.Now()

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create session dir: %w", err)
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	path := filepath.Join(dir, s.ID+".json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write session: %w", err)
	}
	return nil
}

// MarkComplete marks the session as finished and saves it.
func (s *Session) MarkComplete(dir string) error {
	s.Complete = true
	return s.Save(dir)
}

// LoadSession reads a specific session by ID from dir.
func LoadSession(dir, id string) (*Session, error) {
	data, err := os.ReadFile(filepath.Join(dir, id+".json"))
	if err != nil {
		return nil, err
	}
	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}
	return &s, nil
}

// FindIncomplete returns the most recent incomplete session in dir,
// or nil if none exists.
func FindIncomplete(dir string) *Session {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() > entries[j].Name()
	})

	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var s Session
		if err := json.Unmarshal(data, &s); err != nil {
			continue
		}
		if !s.Complete {
			return &s
		}
	}
	return nil
}
