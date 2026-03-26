package face

import "github.com/charmbracelet/lipgloss"

// Styles holds the lipgloss styles used by the Chat viewport.
type Styles struct {
	User       lipgloss.Style
	Agent      lipgloss.Style
	AgentLabel lipgloss.Style
	Tool       lipgloss.Style
	Action     lipgloss.Style
	Approved   lipgloss.Style
	Rejected   lipgloss.Style
	Status     lipgloss.Style
	Model      lipgloss.Style
}

// DefaultStyles returns the standard colour palette shared across
// axon-face applications.
func DefaultStyles() Styles {
	return Styles{
		User: lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")).
			Bold(true),
		Agent: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		AgentLabel: lipgloss.NewStyle().
			Foreground(lipgloss.Color("5")).
			Bold(true),
		Tool: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true),
		Action: lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Bold(true),
		Approved: lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Bold(true),
		Rejected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true),
		Status: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")),
		Model: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true),
	}
}
