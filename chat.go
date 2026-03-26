package face

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	loop "github.com/benaskins/axon-loop"
	talk "github.com/benaskins/axon-loop" // Message is aliased from axon-talk
	tool "github.com/benaskins/axon-tool"
)

// Chat is an embeddable conversational TUI component. It manages a
// viewport, textarea, streaming state, and conversation history.
// It does not implement tea.Model — the outer model calls its methods.
type Chat struct {
	Entries   []Entry
	Streaming string
	Waiting   bool

	Input    textarea.Model
	Viewport viewport.Model
	Messages []talk.Message

	Width  int
	Height int
	Ready  bool

	AgentName string
	Styles    Styles
}

// New creates a Chat with sensible defaults.
func New(agentName string) Chat {
	ta := textarea.New()
	ta.Placeholder = "Type your response..."
	ta.Focus()
	ta.CharLimit = 0
	ta.SetHeight(3)
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)

	return Chat{
		AgentName: agentName,
		Styles:    DefaultStyles(),
		Input:     ta,
	}
}

// InitCmd returns the textarea blink command for use in the outer model's Init().
func (c *Chat) InitCmd() tea.Cmd {
	return textarea.Blink
}

// HandleKey processes base key events (enter, tab, ctrl+c).
// Returns a command and whether the key was handled.
// On enter: appends the user message, sets Waiting=true, and returns
// handled=true with a nil command. The outer model should call StartStream.
func (c *Chat) HandleKey(msg tea.KeyMsg) (cmd tea.Cmd, handled bool) {
	switch msg.String() {
	case "ctrl+c":
		return tea.Quit, true
	case "enter":
		if c.Waiting {
			return nil, true
		}
		text := strings.TrimSpace(c.Input.Value())
		if text == "" {
			return nil, true
		}
		c.Input.Reset()
		c.SendUser(text)
		return nil, true
	case "tab":
		c.toggleLastToolEntry()
		c.RefreshViewport()
		return nil, true
	}
	return nil, false
}

// HandleStreamTick processes a streaming event from the LLM.
// Returns a command to continue reading from the stream, or nil when done.
func (c *Chat) HandleStreamTick(msg StreamTickMsg) tea.Cmd {
	ev := msg.Event
	if ev.Err != nil {
		slog.Error("LLM error", "error", ev.Err)
		c.Entries = append(c.Entries, Entry{Role: RoleAgent, Content: fmt.Sprintf("Error: %v", ev.Err)})
		c.Streaming = ""
		c.Waiting = false
		c.RefreshViewport()
		return nil
	}
	if ev.Tool != nil {
		slog.Info("tool use", "tool", ev.Tool.Name, "args", ev.Tool.Args)
		label := fmt.Sprintf("\u21b3 %s", ev.Tool.Name)
		if len(ev.Tool.Args) > 0 {
			var parts []string
			for k, v := range ev.Tool.Args {
				parts = append(parts, fmt.Sprintf("%s=%v", k, v))
			}
			label += " " + strings.Join(parts, ", ")
		}
		c.Entries = append(c.Entries, Entry{Role: RoleTool, Content: label, Collapsed: true})
		c.RefreshViewport()
		return WaitForEvent(msg.Ch)
	}
	if ev.Done {
		content := ev.Content
		if content == "" {
			content = c.Streaming
		}
		if content != "" {
			slog.Info("agent response", "length", len(content))
			c.Entries = append(c.Entries, Entry{Role: RoleAgent, Content: content})
			c.Messages = append(c.Messages, talk.Message{Role: talk.RoleAssistant, Content: content})
		}
		c.Streaming = ""
		c.Waiting = false
		c.RefreshViewport()
		return nil
	}
	if ev.Token != "" {
		c.Streaming += ev.Token
		c.RefreshViewport()
	}
	return WaitForEvent(msg.Ch)
}

// HandleResize updates dimensions and re-renders the viewport.
func (c *Chat) HandleResize(msg tea.WindowSizeMsg) {
	c.Width = msg.Width
	c.Height = msg.Height
	inputHeight := 5
	statusHeight := 1

	if !c.Ready {
		c.Viewport = viewport.New(msg.Width, msg.Height-inputHeight-statusHeight)
		c.Viewport.YPosition = 0
		c.Ready = true
	} else {
		c.Viewport.Width = msg.Width
		c.Viewport.Height = msg.Height - inputHeight - statusHeight
	}
	c.Input.SetWidth(msg.Width - 2)
	c.RefreshViewport()
}

// SendUser appends a user message to both entries and LLM history,
// and sets the chat to waiting state.
func (c *Chat) SendUser(content string) {
	slog.Info("user message", "length", len(content))
	c.Entries = append(c.Entries, Entry{Role: RoleUser, Content: content})
	c.Messages = append(c.Messages, talk.Message{Role: talk.RoleUser, Content: content})
	c.Waiting = true
	c.Streaming = ""
	c.RefreshViewport()
}

// AppendEntry adds a display-only entry (no effect on LLM history).
func (c *Chat) AppendEntry(e Entry) {
	c.Entries = append(c.Entries, e)
	c.RefreshViewport()
}

// StartStream launches an LLM conversation via loop.Stream and returns
// a Bubble Tea command that feeds events back through StreamTickMsg.
func (c *Chat) StartStream(client talk.LLMClient, req *talk.Request, tools map[string]tool.ToolDef) tea.Cmd {
	messages := make([]talk.Message, len(c.Messages))
	copy(messages, c.Messages)
	req.Messages = messages
	req.Stream = true

	toolDefs := make([]tool.ToolDef, 0, len(tools))
	for _, t := range tools {
		toolDefs = append(toolDefs, t)
	}
	req.Tools = toolDefs

	cfg := loop.RunConfig{
		Client:  client,
		Request: req,
		Tools:   tools,
		ToolCtx: &tool.ToolContext{Ctx: context.Background()},
	}

	ch := loop.Stream(context.Background(), cfg)
	return WaitForEvent(ch)
}

// RefreshViewport rebuilds the viewport content from entries and streaming state.
func (c *Chat) RefreshViewport() {
	var sb strings.Builder
	w := c.Width
	if w <= 0 {
		w = 80
	}
	contentWidth := w - 2
	if contentWidth < 20 {
		contentWidth = 20
	}

	labelWidth := len(c.AgentName) + 2 // label + space padding

	for _, e := range c.Entries {
		switch e.Role {
		case RoleUser:
			sb.WriteString(c.Styles.User.Render("you") + " ")
			sb.WriteString(WordWrap(e.Content, contentWidth-5))
			sb.WriteString("\n\n")
		case RoleAgent:
			sb.WriteString(c.Styles.AgentLabel.Render(c.AgentName) + " ")
			sb.WriteString(c.Styles.Agent.Width(contentWidth-labelWidth).Render(e.Content))
			sb.WriteString("\n\n")
		case RoleTool:
			sb.WriteString(c.Styles.Tool.Render(e.Content))
			sb.WriteString("\n")
		case RoleAction:
			sb.WriteString(c.Styles.Action.Render(e.Content))
			sb.WriteString("\n")
		}
	}

	if c.Streaming != "" {
		sb.WriteString(c.Styles.AgentLabel.Render(c.AgentName) + " ")
		sb.WriteString(c.Styles.Agent.Width(contentWidth-labelWidth).Render(c.Streaming))
		sb.WriteString("\u2588")
		sb.WriteString("\n")
	}

	c.Viewport.SetContent(sb.String())
	c.Viewport.GotoBottom()
}

// View returns the composed layout: viewport + status bar + input.
// The caller provides the status text (e.g. "thinking..." or key hints).
func (c *Chat) View(status string) string {
	if !c.Ready {
		return "Initializing..."
	}
	return lipgloss.JoinVertical(
		lipgloss.Left,
		c.Viewport.View(),
		status,
		c.Input.View(),
	)
}

// UpdateInput forwards a message to the textarea and viewport sub-components.
// Call this in the outer model's Update for messages not handled by HandleKey.
func (c *Chat) UpdateInput(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	c.Input, cmd = c.Input.Update(msg)
	cmds = append(cmds, cmd)
	c.Viewport, cmd = c.Viewport.Update(msg)
	cmds = append(cmds, cmd)
	return tea.Batch(cmds...)
}

func (c *Chat) toggleLastToolEntry() {
	for i := len(c.Entries) - 1; i >= 0; i-- {
		if c.Entries[i].Role == RoleTool {
			c.Entries[i].Collapsed = !c.Entries[i].Collapsed
			return
		}
	}
}
