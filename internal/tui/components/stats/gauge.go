package stats

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/tzy0608/leet-tui/internal/tui/theme"
)

var (
	labelStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#6C7086", Dark: "#6C7086"}).
		Width(8)

	valueStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#4C4F69", Dark: "#CDD6F4"}).
		Bold(true)

	barEmptyStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#E6E9EF", Dark: "#313244"})
)

// ProgressBar renders a horizontal progress bar using thin line characters.
func ProgressBar(current, total int, width int) string {
	if total == 0 {
		return barEmptyStyle.Render(strings.Repeat("─", width))
	}

	ratio := math.Min(1.0, float64(current)/float64(total))
	filled := int(ratio * float64(width))
	empty := width - filled

	filledStyle := lipgloss.NewStyle().Foreground(theme.Current.Primary)
	bar := filledStyle.Render(strings.Repeat("━", filled)) +
		barEmptyStyle.Render(strings.Repeat("─", empty))
	return bar
}

// ColoredProgressBar renders a progress bar with custom color for difficulty levels.
func ColoredProgressBar(current, total, width int, color lipgloss.AdaptiveColor) string {
	if total == 0 {
		return barEmptyStyle.Render(strings.Repeat("─", width))
	}

	ratio := math.Min(1.0, float64(current)/float64(total))
	filled := int(ratio * float64(width))
	empty := width - filled

	filledStyle := lipgloss.NewStyle().Foreground(color)
	bar := filledStyle.Render(strings.Repeat("━", filled)) +
		barEmptyStyle.Render(strings.Repeat("─", empty))
	return bar
}

// StatRow renders a single stat line with label, bar, and value.
func StatRow(label string, current, total int, barWidth int) string {
	bar := ProgressBar(current, total, barWidth)
	pct := 0
	if total > 0 {
		pct = int(float64(current) / float64(total) * 100)
	}
	val := valueStyle.Width(10).Align(lipgloss.Right).Render(fmt.Sprintf("%d/%d (%d%%)", current, total, pct))
	return lipgloss.NewStyle().Padding(0, 1).Render(labelStyle.Render(label) + " " + bar + " " + val)
}

// DifficultyRow renders a compact difficulty stat.
func DifficultyRow(difficulty string, solved, total int, barWidth int) string {
	var color lipgloss.AdaptiveColor
	switch difficulty {
	case "Easy":
		color = theme.Current.Easy
	case "Medium":
		color = theme.Current.Medium
	case "Hard":
		color = theme.Current.Hard
	default:
		color = theme.Current.TextMuted
	}

	style := lipgloss.NewStyle().Foreground(color)
	label := style.Width(7).Render(difficulty[:1] + ":")
	bar := ColoredProgressBar(solved, total, barWidth, color)
	val := valueStyle.Width(10).Align(lipgloss.Right).Render(fmt.Sprintf("%d", solved))
	return lipgloss.NewStyle().Padding(0, 1).Render(label + bar + " " + val)
}
