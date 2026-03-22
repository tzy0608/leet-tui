package styles

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/leet-tui/leet-tui/internal/tui/theme"
)

// Shared styles used across components.
var (
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Current.Primary)

	Subtitle = lipgloss.NewStyle().
		Foreground(theme.Current.TextMuted)

	Border = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current.Border)

	FocusedBorder = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current.Primary)

	// Panel is a generic panel with padding and rounded border.
	Panel = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current.Border).
		Padding(1, 2)

	// FocusedPanel is a panel with Primary-colored border.
	FocusedPanel = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current.Primary).
		Padding(1, 2)

	// PanelHeader is a bold primary-colored header with underline and bottom margin.
	PanelHeader = lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Current.Primary).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(theme.Current.Primary).
		MarginBottom(1)

	// Card is a dashboard card with subtle background and padding.
	Card = lipgloss.NewStyle().
		Background(lipgloss.AdaptiveColor{Light: "#EFF1F5", Dark: "#1E1E2E"}).
		Padding(1, 2)

	// RowEven is the even-row style with padding.
	RowEven = lipgloss.NewStyle().
		Padding(0, 1)

	// RowOdd is the odd-row style with subtle background and padding.
	RowOdd = lipgloss.NewStyle().
		Background(lipgloss.AdaptiveColor{Light: "#E6E9EF", Dark: "#181825"}).
		Padding(0, 1)

	// HintKey styles shortcut keys with accent color.
	HintKey = lipgloss.NewStyle().
		Foreground(theme.Current.Accent).
		Bold(true)

	// HintSep is a muted separator for hints.
	HintSep = lipgloss.NewStyle().
		Foreground(theme.Current.TextMuted)

	StatusBar = lipgloss.NewStyle().
		Foreground(theme.Current.Text).
		Background(lipgloss.AdaptiveColor{Light: "#E6E9EF", Dark: "#313244"}).
		Padding(0, 1)

	AcceptedBadge = lipgloss.NewStyle().
		Foreground(theme.Current.Success).
		Bold(true)

	DifficultyEasy = lipgloss.NewStyle().
		Foreground(theme.Current.Easy)

	DifficultyMedium = lipgloss.NewStyle().
		Foreground(theme.Current.Medium)

	DifficultyHard = lipgloss.NewStyle().
		Foreground(theme.Current.Hard)

	TabActive = lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Current.Primary).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(theme.Current.Primary)

	TabInactive = lipgloss.NewStyle().
		Foreground(theme.Current.TextMuted)

	HelpKey = lipgloss.NewStyle().
		Foreground(theme.Current.Accent).
		Bold(true)

	HelpDesc = lipgloss.NewStyle().
		Foreground(theme.Current.TextMuted)
)

// NewBadge creates an inline badge with the given foreground and background colors.
func NewBadge(fg, bg lipgloss.AdaptiveColor) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(fg).
		Background(bg).
		Bold(true).
		Padding(0, 1)
}

// VerticalDivider returns a vertical line of │ characters with the border color.
func VerticalDivider(height int) string {
	style := lipgloss.NewStyle().Foreground(theme.Current.Border)
	lines := make([]string, height)
	for i := range lines {
		lines[i] = style.Render("│")
	}
	return strings.Join(lines, "\n")
}

// DifficultyStyle returns the appropriate style for a difficulty level.
func DifficultyStyle(difficulty string) lipgloss.Style {
	switch difficulty {
	case "Easy":
		return DifficultyEasy
	case "Medium":
		return DifficultyMedium
	case "Hard":
		return DifficultyHard
	default:
		return Subtitle
	}
}
