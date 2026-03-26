// Package face provides reusable Bubble Tea components for building
// LLM-powered conversational TUIs on top of axon-loop.
package face

// Role constants for conversation entries.
const (
	RoleUser   = "user"
	RoleAgent  = "agent"
	RoleTool   = "tool"
	RoleAction = "action"
)

// Entry is a single item in the conversation view.
type Entry struct {
	Role      string
	Content   string
	Collapsed bool // tool entries can be toggled
}
