package face

import (
	"testing"

	loop "github.com/benaskins/axon-loop"
	tea "github.com/charmbracelet/bubbletea"
)

func TestNewChat(t *testing.T) {
	c := New("vita")
	if c.AgentName != "vita" {
		t.Fatalf("AgentName = %q, want %q", c.AgentName, "vita")
	}
	if c.Waiting {
		t.Fatal("new chat should not be waiting")
	}
	if len(c.Entries) != 0 {
		t.Fatalf("new chat should have no entries, got %d", len(c.Entries))
	}
}

func TestSendUser(t *testing.T) {
	c := New("test")
	c.SendUser("hello")

	if len(c.Entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(c.Entries))
	}
	if c.Entries[0].Role != RoleUser {
		t.Fatalf("role = %q, want %q", c.Entries[0].Role, RoleUser)
	}
	if c.Entries[0].Content != "hello" {
		t.Fatalf("content = %q, want %q", c.Entries[0].Content, "hello")
	}
	if len(c.Messages) != 1 {
		t.Fatalf("messages = %d, want 1", len(c.Messages))
	}
	if !c.Waiting {
		t.Fatal("should be waiting after SendUser")
	}
}

func TestHandleKeyEnter(t *testing.T) {
	c := New("test")
	c.Input.SetValue("hello world")

	cmd, handled := c.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if !handled {
		t.Fatal("enter should be handled")
	}
	if cmd != nil {
		t.Fatal("enter should return nil cmd (app calls StartStream)")
	}
	if len(c.Entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(c.Entries))
	}
	if c.Input.Value() != "" {
		t.Fatal("input should be reset after enter")
	}
}

func TestHandleKeyEnterWhileWaiting(t *testing.T) {
	c := New("test")
	c.Waiting = true
	c.Input.SetValue("ignored")

	_, handled := c.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if !handled {
		t.Fatal("enter while waiting should be handled (swallowed)")
	}
	if len(c.Entries) != 0 {
		t.Fatal("should not append entry while waiting")
	}
}

func TestHandleStreamTickDone(t *testing.T) {
	c := New("test")
	c.Waiting = true
	c.Streaming = "partial"

	cmd := c.HandleStreamTick(StreamTickMsg{
		Event: StreamEvent{Done: true, Content: "full response"},
	})
	if cmd != nil {
		t.Fatal("done event should return nil cmd")
	}
	if c.Waiting {
		t.Fatal("should not be waiting after done")
	}
	if c.Streaming != "" {
		t.Fatal("streaming should be cleared after done")
	}
	if len(c.Entries) != 1 || c.Entries[0].Content != "full response" {
		t.Fatalf("expected agent entry with full response, got %v", c.Entries)
	}
	if len(c.Messages) != 1 || c.Messages[0].Role != loop.RoleAssistant {
		t.Fatal("should append assistant message to history")
	}
}

func TestHandleStreamTickToken(t *testing.T) {
	c := New("test")
	c.Waiting = true

	ch := make(chan loop.Event, 1)
	ch <- loop.Event{Token: "world"}

	cmd := c.HandleStreamTick(StreamTickMsg{
		Event: StreamEvent{Token: "hello "},
		Ch:    ch,
	})
	if c.Streaming != "hello " {
		t.Fatalf("streaming = %q, want %q", c.Streaming, "hello ")
	}
	if cmd == nil {
		t.Fatal("token event should return cmd to read next event")
	}
}

func TestHandleStreamTickToolUse(t *testing.T) {
	c := New("test")
	c.Waiting = true

	ch := make(chan loop.Event)
	go func() { close(ch) }()

	cmd := c.HandleStreamTick(StreamTickMsg{
		Event: StreamEvent{Tool: &ToolUseEvent{Name: "search", Args: map[string]any{"q": "test"}}},
		Ch:    ch,
	})
	if len(c.Entries) != 1 || c.Entries[0].Role != RoleTool {
		t.Fatal("should append tool entry")
	}
	if cmd == nil {
		t.Fatal("tool event should return cmd to read next event")
	}
}

func TestWordWrap(t *testing.T) {
	got := WordWrap("hello world foo bar", 10)
	if got == "hello world foo bar" {
		t.Fatal("should have wrapped")
	}
}
