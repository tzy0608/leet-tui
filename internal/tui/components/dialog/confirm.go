package dialog

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leet-tui/leet-tui/internal/tui/theme"
)

// ConfirmMsg is sent when the user makes a choice.
type ConfirmMsg struct {
	Confirmed bool
}

// Confirm is a yes/no confirmation dialog.
type Confirm struct {
	title   string
	message string
	cursor  int // 0=Yes, 1=No
}

var (
	dialogStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current.Primary).
		Padding(2, 4).
		Width(40)

	dialogTitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Current.Primary).
		MarginBottom(1)

	dialogMsg = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#4C4F69", Dark: "#CDD6F4"}).
		MarginBottom(1)

	buttonActive = lipgloss.NewStyle().
		Background(theme.Current.Primary).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 2).
		MarginRight(2)

	buttonInactive = lipgloss.NewStyle().
		Background(lipgloss.AdaptiveColor{Light: "#E6E9EF", Dark: "#313244"}).
		Foreground(lipgloss.AdaptiveColor{Light: "#6C7086", Dark: "#6C7086"}).
		Padding(0, 2).
		MarginRight(2)
)

// NewConfirm creates a new confirmation dialog.
func NewConfirm(title, message string) *Confirm {
	return &Confirm{title: title, message: message}
}

// Update handles keyboard input.
func (c *Confirm) Update(msg tea.Msg) (*Confirm, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "left", "h", "tab":
			c.cursor = 0
		case "right", "l":
			c.cursor = 1
		case "enter":
			return c, func() tea.Msg { return ConfirmMsg{Confirmed: c.cursor == 0} }
		case "y", "Y":
			return c, func() tea.Msg { return ConfirmMsg{Confirmed: true} }
		case "n", "N", "esc":
			return c, func() tea.Msg { return ConfirmMsg{Confirmed: false} }
		}
	}
	return c, nil
}

// View renders the confirmation dialog.
func (c *Confirm) View() string {
	var yes, no lipgloss.Style
	if c.cursor == 0 {
		yes = buttonActive
		no = buttonInactive
	} else {
		yes = buttonInactive
		no = buttonActive
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Top,
		yes.Render("Yes"),
		no.Render("No"),
	)

	content := dialogTitle.Render(c.title) + "\n" +
		dialogMsg.Render(c.message) + "\n" +
		buttons

	return dialogStyle.Render(content)
}
