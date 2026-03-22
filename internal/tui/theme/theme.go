package theme

import "github.com/charmbracelet/lipgloss"

// Colors defines the color palette for the TUI.
type Colors struct {
	Primary    lipgloss.AdaptiveColor
	Secondary  lipgloss.AdaptiveColor
	Accent     lipgloss.AdaptiveColor
	Text       lipgloss.AdaptiveColor
	TextMuted  lipgloss.AdaptiveColor
	Border     lipgloss.AdaptiveColor
	Error      lipgloss.AdaptiveColor
	Success    lipgloss.AdaptiveColor
	Warning    lipgloss.AdaptiveColor
	Easy       lipgloss.AdaptiveColor
	Medium     lipgloss.AdaptiveColor
	Hard       lipgloss.AdaptiveColor
}

// Current holds the active theme colors.
var Current = DefaultColors()

// DefaultColors returns the default Catppuccin-inspired color scheme.
func DefaultColors() Colors {
	return Colors{
		Primary:   lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: "#7C3AED"},
		Secondary: lipgloss.AdaptiveColor{Light: "#06B6D4", Dark: "#06B6D4"},
		Accent:    lipgloss.AdaptiveColor{Light: "#F59E0B", Dark: "#F59E0B"},
		Text:      lipgloss.AdaptiveColor{Light: "#1E1E2E", Dark: "#CDD6F4"},
		TextMuted: lipgloss.AdaptiveColor{Light: "#6C7086", Dark: "#6C7086"},
		Border:    lipgloss.AdaptiveColor{Light: "#9399B2", Dark: "#45475A"},
		Error:     lipgloss.AdaptiveColor{Light: "#D20F39", Dark: "#F38BA8"},
		Success:   lipgloss.AdaptiveColor{Light: "#40A02B", Dark: "#A6E3A1"},
		Warning:   lipgloss.AdaptiveColor{Light: "#DF8E1D", Dark: "#F9E2AF"},
		Easy:      lipgloss.AdaptiveColor{Light: "#40A02B", Dark: "#A6E3A1"},
		Medium:    lipgloss.AdaptiveColor{Light: "#DF8E1D", Dark: "#F9E2AF"},
		Hard:      lipgloss.AdaptiveColor{Light: "#D20F39", Dark: "#F38BA8"},
	}
}

// DifficultyColor returns the color for a difficulty level.
func DifficultyColor(difficulty string) lipgloss.AdaptiveColor {
	switch difficulty {
	case "Easy":
		return Current.Easy
	case "Medium":
		return Current.Medium
	case "Hard":
		return Current.Hard
	default:
		return Current.TextMuted
	}
}
