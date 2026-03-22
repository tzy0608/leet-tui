package layout

import (
	"github.com/charmbracelet/lipgloss"
)

// Container wraps content with border and padding.
type Container struct {
	Title  string
	Style  lipgloss.Style
	Width  int
	Height int
}

// NewContainer creates a bordered container.
func NewContainer(title string) *Container {
	return &Container{
		Title: title,
		Style: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#6C7086")).
			Padding(0, 1),
	}
}

// Render wraps content in a bordered container.
func (c *Container) Render(content string) string {
	style := c.Style.
		Width(c.Width - 2).  // account for border
		Height(c.Height - 2) // account for border

	if c.Title != "" {
		style = style.BorderTop(true)
		titleStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CDD6F4")).
			Bold(true)
		title := titleStyle.Render(" " + c.Title + " ")
		return title + "\n" + style.Render(content)
	}

	return style.Render(content)
}

// SetSize updates container dimensions.
func (c *Container) SetSize(w, h int) {
	c.Width = w
	c.Height = h
}
