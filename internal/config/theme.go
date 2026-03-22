package config

import "github.com/charmbracelet/lipgloss"

type ThemeColors struct {
	Primary    lipgloss.Color
	Secondary  lipgloss.Color
	Accent     lipgloss.Color
	Background lipgloss.Color
	Foreground lipgloss.Color
	Muted      lipgloss.Color
	Error      lipgloss.Color
	Success    lipgloss.Color
	Warning    lipgloss.Color
	Easy       lipgloss.Color
	Medium     lipgloss.Color
	Hard       lipgloss.Color
}

var DefaultTheme = ThemeColors{
	Primary:    lipgloss.Color("#7C3AED"),
	Secondary:  lipgloss.Color("#06B6D4"),
	Accent:     lipgloss.Color("#F59E0B"),
	Background: lipgloss.Color("#1E1E2E"),
	Foreground: lipgloss.Color("#CDD6F4"),
	Muted:      lipgloss.Color("#6C7086"),
	Error:      lipgloss.Color("#F38BA8"),
	Success:    lipgloss.Color("#A6E3A1"),
	Warning:    lipgloss.Color("#F9E2AF"),
	Easy:       lipgloss.Color("#A6E3A1"),
	Medium:     lipgloss.Color("#F9E2AF"),
	Hard:       lipgloss.Color("#F38BA8"),
}
