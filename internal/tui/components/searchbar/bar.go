package searchbar

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tzy0608/leet-tui/internal/tui/theme"
)

// SearchMsg is sent when the search query changes.
type SearchMsg struct {
	Query string
}

// Bar is a search input component.
type Bar struct {
	input     textinput.Model
	focused   bool
	width     int
	prevValue string
}

var (
	containerStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(lipgloss.AdaptiveColor{Light: "#CCD0DA", Dark: "#45475A"}).
		Padding(0, 2)

	focusedContainerStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(theme.Current.Primary).
		Background(lipgloss.AdaptiveColor{Light: "#EFF1F5", Dark: "#1E1E2E"}).
		Padding(0, 2)

	slashStyle = lipgloss.NewStyle().
		Foreground(theme.Current.Accent).
		Bold(true)
)

// New creates a new search bar.
func New() *Bar {
	ti := textinput.New()
	ti.Placeholder = "Search problems..."
	ti.CharLimit = 100
	ti.PromptStyle = lipgloss.NewStyle().Foreground(theme.Current.Primary)
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#4C4F69", Dark: "#CDD6F4"})

	return &Bar{input: ti}
}

// Focus focuses the search bar.
func (b *Bar) Focus() tea.Cmd {
	b.focused = true
	return b.input.Focus()
}

// Blur unfocuses the search bar.
func (b *Bar) Blur() tea.Cmd {
	b.focused = false
	b.input.Blur()
	return nil
}

// Value returns the current search query.
func (b *Bar) Value() string {
	return b.input.Value()
}

// Clear resets the search query.
func (b *Bar) Clear() {
	b.input.Reset()
}

// SetWidth updates the bar width.
func (b *Bar) SetWidth(w int) {
	b.width = w
	b.input.Width = w - 6 // account for padding and prompt
}

// Update handles messages.
func (b *Bar) Update(msg tea.Msg) (*Bar, tea.Cmd) {
	var cmd tea.Cmd
	b.input, cmd = b.input.Update(msg)

	// Only emit SearchMsg when the value actually changes
	if _, ok := msg.(tea.KeyMsg); ok {
		current := b.input.Value()
		if current != b.prevValue {
			b.prevValue = current
			return b, tea.Batch(cmd, func() tea.Msg {
				return SearchMsg{Query: current}
			})
		}
	}

	return b, cmd
}

// View renders the search bar.
func (b *Bar) View() string {
	style := containerStyle
	if b.focused {
		style = focusedContainerStyle
	}
	prefix := slashStyle.Render("/") + " "
	return style.Width(b.width - 2).Render(prefix + b.input.View())
}
