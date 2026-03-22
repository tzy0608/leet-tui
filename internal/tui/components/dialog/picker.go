package dialog

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leet-tui/leet-tui/internal/tui/theme"
)

// PickedMsg is sent when the user selects an item.
type PickedMsg struct {
	Value string
}

// Picker is a simple item selection dialog.
type Picker struct {
	title  string
	items  []string
	cursor int
}

// NewPicker creates a new picker dialog.
func NewPicker(title string, items []string) *Picker {
	return &Picker{title: title, items: items}
}

// Update handles keyboard input.
func (p *Picker) Update(msg tea.Msg) (*Picker, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "up", "k":
			if p.cursor > 0 {
				p.cursor--
			}
		case "down", "j":
			if p.cursor < len(p.items)-1 {
				p.cursor++
			}
		case "enter":
			if p.cursor < len(p.items) {
				val := p.items[p.cursor]
				return p, func() tea.Msg { return PickedMsg{Value: val} }
			}
		case "esc":
			return p, func() tea.Msg { return PickedMsg{Value: ""} }
		}
	}
	return p, nil
}

// View renders the picker dialog.
func (p *Picker) View() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Current.Primary).
		MarginBottom(1).
		Render(p.title)

	selectedStyle := lipgloss.NewStyle().
		Background(theme.Current.Primary).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1)

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#4C4F69", Dark: "#CDD6F4"}).
		Padding(0, 1)

	var rows []string
	rows = append(rows, title)
	for i, item := range p.items {
		if i == p.cursor {
			rows = append(rows, selectedStyle.Render("▸ "+item))
		} else {
			rows = append(rows, normalStyle.Render("  "+item))
		}
	}

	content := ""
	for i, r := range rows {
		if i > 0 {
			content += "\n"
		}
		content += r
	}

	return dialogStyle.Width(36).Render(content)
}
