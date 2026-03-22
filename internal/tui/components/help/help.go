package help

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/leet-tui/leet-tui/internal/tui/styles"
	"github.com/leet-tui/leet-tui/internal/tui/theme"
)

// Overlay renders a key binding help overlay.
type Overlay struct {
	bindings []key.Binding
	width    int
	height   int
}

var (
	overlayStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current.Primary).
		Padding(1, 2)

	keyBadgeStyle = styles.NewBadge(
		lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#1E1E2E"},
		theme.Current.Accent,
	)

	descStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#4C4F69", Dark: "#CDD6F4"})

	helpTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Current.Primary).
		MarginBottom(1)
)

// New creates a help overlay with the given bindings.
func New(bindings []key.Binding) *Overlay {
	return &Overlay{bindings: bindings}
}

// SetSize updates the overlay dimensions.
func (o *Overlay) SetSize(w, h int) {
	o.width = w
	o.height = h
}

// View renders the help overlay.
func (o *Overlay) View() string {
	var lines []string
	lines = append(lines, helpTitleStyle.Render("Keyboard Shortcuts"))

	for _, b := range o.bindings {
		if b.Help().Key == "" {
			continue
		}
		line := lipgloss.JoinHorizontal(lipgloss.Top,
			keyBadgeStyle.Render(b.Help().Key),
			"  ",
			descStyle.Render(b.Help().Desc),
		)
		lines = append(lines, line)
	}

	content := strings.Join(lines, "\n")
	maxW := o.width - 10
	if maxW < 40 {
		maxW = 40
	}

	return overlayStyle.Width(maxW).Render(content)
}

// GlobalBindings returns the global key bindings shown on all pages.
func GlobalBindings() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "Dashboard")),
		key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "Problems")),
		key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "Review")),
		key.NewBinding(key.WithKeys("4"), key.WithHelp("4", "Plans")),
		key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "Toggle help")),
		key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "Quit")),
	}
}

// FormatShortBindings formats a compact single-line help string.
func FormatShortBindings(bindings []key.Binding) string {
	sep := styles.HintSep.Render(" │ ")
	var parts []string
	for _, b := range bindings {
		h := b.Help()
		if h.Key == "" {
			continue
		}
		parts = append(parts, styles.HintKey.Render(h.Key)+":"+h.Desc)
	}
	return strings.Join(parts, sep)
}
