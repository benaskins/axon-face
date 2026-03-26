package face

import (
	loop "github.com/benaskins/axon-loop"
	tea "github.com/charmbracelet/bubbletea"
)

// ToolUseEvent carries tool invocation details from the LLM.
type ToolUseEvent struct {
	Name string
	Args map[string]any
}

// StreamEvent is a parsed event from the LLM stream.
type StreamEvent struct {
	Token   string
	Tool    *ToolUseEvent
	Done    bool
	Err     error
	Content string // final content on done
}

// StreamTickMsg wraps a stream event for the Bubble Tea update loop.
// The channel carries the next event for re-subscription.
type StreamTickMsg struct {
	Event StreamEvent
	Ch    <-chan loop.Event
}

// WaitForEvent reads the next loop.Event from the stream channel and
// converts it to a StreamTickMsg for Bubble Tea's update loop.
func WaitForEvent(ch <-chan loop.Event) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		if !ok {
			return StreamTickMsg{Event: StreamEvent{Done: true}}
		}
		se := StreamEvent{}
		switch {
		case ev.Err != nil:
			se.Err = ev.Err
		case ev.ToolUse != nil:
			se.Tool = &ToolUseEvent{Name: ev.ToolUse.Name, Args: ev.ToolUse.Args}
		case ev.Done != nil:
			se.Done = true
			se.Content = ev.Done.Content
		case ev.Token != "":
			se.Token = ev.Token
		}
		return StreamTickMsg{Event: se, Ch: ch}
	}
}
