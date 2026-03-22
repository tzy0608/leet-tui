package page

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
)

// ID identifies a page.
type ID int

const (
	DashboardPage ID = iota
	ProblemsPage
	ProblemDetailPage
	ReviewPage
	PlansPage
)

func (id ID) String() string {
	switch id {
	case DashboardPage:
		return "Dashboard"
	case ProblemsPage:
		return "Problems"
	case ProblemDetailPage:
		return "Problem Detail"
	case ReviewPage:
		return "Review"
	case PlansPage:
		return "Plans"
	default:
		return "Unknown"
	}
}

// ChangeMsg requests a page navigation.
type ChangeMsg struct {
	Target    ID
	Args      map[string]any
}

// InputFocuser is implemented by pages that have a text input which should
// suppress global navigation keys when focused.
type InputFocuser interface {
	IsInputFocused() bool
}

// ErrorMsg reports errors from data loading to be shown in the status bar.
type ErrorMsg struct {
	Err error
}

// NavigateTo creates a command to navigate to a page.
func NavigateTo(target ID, args ...map[string]any) tea.Cmd {
	return func() tea.Msg {
		msg := ChangeMsg{Target: target}
		if len(args) > 0 {
			msg.Args = args[0]
		}
		return msg
	}
}

// renderMarkdown renders markdown content with glamour, falling back to raw text on error.
func renderMarkdown(md string, width int) string {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return md
	}
	rendered, err := renderer.Render(md)
	if err != nil {
		return md
	}
	return strings.TrimRight(rendered, "\n")
}
