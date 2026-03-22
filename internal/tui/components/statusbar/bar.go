package statusbar

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/tzy0608/leet-tui/internal/tui/styles"
	"github.com/tzy0608/leet-tui/internal/tui/theme"
)

// Bar is the bottom status bar component.
type Bar struct {
	width   int
	page    string
	lang    string
	status  string
	keyHint string
}

var (
	barStyle = lipgloss.NewStyle().
		Background(lipgloss.AdaptiveColor{Light: "#E6E9EF", Dark: "#313244"})

	pageStyle = lipgloss.NewStyle().
		Background(theme.Current.Primary).
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true).
		Padding(0, 1)

	langStyle = lipgloss.NewStyle().
		Background(theme.Current.Secondary).
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true).
		Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#6C7086", Dark: "#6C7086"}).
		Padding(0, 1)

	sepStyle = styles.HintSep

	hintTextStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#7C7F93", Dark: "#7F849C"})
)

// New creates a new status bar.
func New() *Bar {
	return &Bar{}
}

// SetPage updates the page indicator.
func (b *Bar) SetPage(page string) {
	b.page = page
}

// SetLang updates the language indicator.
func (b *Bar) SetLang(lang string) {
	b.lang = lang
}

// SetStatus updates the status message.
func (b *Bar) SetStatus(status string) {
	b.status = status
}

// SetWidth updates the bar width.
func (b *Bar) SetWidth(w int) {
	b.width = w
}

func (b *Bar) formatKeyHints() string {
	sep := sepStyle.Render(" │ ")
	hints := styles.HintKey.Render("1") + " Home" + sep +
		styles.HintKey.Render("2") + " Problems" + sep +
		styles.HintKey.Render("3") + " Review" + sep +
		styles.HintKey.Render("4") + " Plans" + sep +
		styles.HintKey.Render("?") + " Help" + sep +
		styles.HintKey.Render("q") + " Quit"
	return hintTextStyle.Render(hints)
}

// View renders the status bar.
func (b *Bar) View() string {
	left := pageStyle.Render(b.page)

	center := ""
	if b.status != "" {
		center = statusStyle.Render(b.status)
	}

	langPart := ""
	if b.lang != "" {
		langPart = langStyle.Render(b.lang) + " "
	}
	right := langPart + b.formatKeyHints()

	// Calculate spacing
	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	centerW := lipgloss.Width(center)
	gap := b.width - leftW - rightW - centerW
	if gap < 0 {
		gap = 0
	}

	spacer := lipgloss.NewStyle().Width(gap).Render("")
	row := lipgloss.JoinHorizontal(lipgloss.Bottom, left, center, spacer, right)

	return barStyle.Width(b.width).Render(row)
}
