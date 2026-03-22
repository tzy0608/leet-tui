package layout

import (
	"github.com/charmbracelet/lipgloss"
)

// SplitPane renders two panes side by side with a configurable ratio.
type SplitPane struct {
	Ratio float64 // 0.0-1.0, left pane ratio
	Width int
	Height int
}

// NewSplitPane creates a horizontal split pane.
func NewSplitPane(ratio float64) *SplitPane {
	return &SplitPane{Ratio: ratio}
}

// SetSize updates the split pane dimensions.
func (s *SplitPane) SetSize(w, h int) {
	s.Width = w
	s.Height = h
}

// LeftWidth returns the width of the left pane.
func (s *SplitPane) LeftWidth() int {
	return int(float64(s.Width) * s.Ratio)
}

// RightWidth returns the width of the right pane.
func (s *SplitPane) RightWidth() int {
	return s.Width - s.LeftWidth()
}

// Render combines left and right content side by side.
func (s *SplitPane) Render(left, right string) string {
	leftStyle := lipgloss.NewStyle().
		Width(s.LeftWidth()).
		Height(s.Height)

	rightStyle := lipgloss.NewStyle().
		Width(s.RightWidth()).
		Height(s.Height)

	return lipgloss.JoinHorizontal(lipgloss.Top,
		leftStyle.Render(left),
		rightStyle.Render(right),
	)
}
