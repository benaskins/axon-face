# axon-face

Reusable [Bubble Tea](https://github.com/charmbracelet/bubbletea) components for building LLM-powered conversational TUIs on top of axon-loop.

Import: `github.com/benaskins/axon-face`

## What it does

axon-face provides an embeddable chat component that handles the full lifecycle of an LLM conversation in the terminal: user input, streaming responses, tool use display, session persistence, and styled rendering via lipgloss.

Consumer apps (imago, vita) compose the `Chat` component into their own Bubble Tea models.

## Components

| Type | Purpose |
|------|---------|
| `Chat` | Embeddable conversational TUI: viewport, textarea, streaming state |
| `Session` | Persisted conversation with messages, phase, and timestamps |
| `Entry` | Single item in the conversation view (user, agent, tool, action) |
| `StreamTickMsg` | Bridges axon-loop events into the Bubble Tea update cycle |
| `Styles` | Configurable lipgloss styles for all conversation elements |

## Usage

```go
import face "github.com/benaskins/axon-face"

chat := face.New("my-agent")
cmd := chat.StartStream(client, req, tools)
```

The Chat component exposes `HandleKey`, `HandleStreamTick`, `HandleResize`, and `View` for integration into a parent Bubble Tea model.

## Session management

Sessions are JSON files that track conversation state across restarts:

```go
session := face.NewSession()
session.Save(dir)

incomplete := face.FindIncomplete(dir)
```

## Dependencies

- axon-loop, axon-talk, axon-tool
- bubbletea, bubbles, lipgloss (Charm)

## Build & Test

```bash
go test ./...
go vet ./...
```
