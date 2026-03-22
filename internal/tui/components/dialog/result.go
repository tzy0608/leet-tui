package dialog

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leet-tui/leet-tui/internal/leetcode"
	"github.com/leet-tui/leet-tui/internal/tui/theme"
)

// ResultDismissMsg is sent when the user closes the result dialog.
type ResultDismissMsg struct{}

// ResultDialog displays submit/run results in a dialog overlay.
type ResultDialog struct {
	title   string
	content string
	loading bool
}

var resultDialogStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(theme.Current.Primary).
	Padding(1, 3).
	Width(56)

var resultTitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(theme.Current.Primary).
	MarginBottom(1)

var resultContentStyle = lipgloss.NewStyle().
	Foreground(lipgloss.AdaptiveColor{Light: "#4C4F69", Dark: "#CDD6F4"})

var resultHintStyle = lipgloss.NewStyle().
	Foreground(lipgloss.AdaptiveColor{Light: "#9CA3AF", Dark: "#585B70"}).
	MarginTop(1)

// NewResultDialog creates a new result dialog in loading state.
func NewResultDialog(title string) *ResultDialog {
	loadingText := "⏳ Submitting..."
	if title == "Run Test" {
		loadingText = "⏳ Running..."
	}
	return &ResultDialog{
		title:   title,
		content: loadingText,
		loading: true,
	}
}

// SetSubmitResult formats a SubmitResult into dialog content.
func (d *ResultDialog) SetSubmitResult(r *leetcode.SubmitResult, err error) {
	d.loading = false
	if err != nil {
		d.content = "❌ Error\n\n" + err.Error()
		return
	}

	var b strings.Builder
	if r.StatusMsg == "Accepted" {
		b.WriteString("✅ Accepted\n")
	} else {
		b.WriteString(fmt.Sprintf("❌ %s\n", r.StatusMsg))
	}

	if r.TotalTestcase > 0 {
		b.WriteString(fmt.Sprintf("Passed: %d/%d\n", r.TotalCorrect, r.TotalTestcase))
	}
	if r.RuntimeMs != "" {
		if r.StatusMsg == "Accepted" && r.RuntimePercentile > 0 {
			b.WriteString(fmt.Sprintf("Runtime: %s  (beats %.2f%%)\n", r.RuntimeMs, r.RuntimePercentile))
		} else {
			b.WriteString(fmt.Sprintf("Runtime: %s\n", r.RuntimeMs))
		}
	}
	if r.MemoryMB != "" {
		if r.StatusMsg == "Accepted" && r.MemoryPercentile > 0 {
			b.WriteString(fmt.Sprintf("Memory: %s  (beats %.2f%%)\n", r.MemoryMB, r.MemoryPercentile))
		} else {
			b.WriteString(fmt.Sprintf("Memory: %s\n", r.MemoryMB))
		}
	}
	if r.CompileError != "" {
		b.WriteString(fmt.Sprintf("\nCompile Error:\n  %s\n", r.CompileError))
	}
	if r.RuntimeError != "" {
		b.WriteString(fmt.Sprintf("\nRuntime Error:\n  %s\n", r.RuntimeError))
	}

	d.content = strings.TrimRight(b.String(), "\n")
}

// SetRunResult formats a RunResult into dialog content.
func (d *ResultDialog) SetRunResult(r *leetcode.RunResult, err error) {
	d.loading = false
	if err != nil {
		d.content = "❌ Error\n\n" + err.Error()
		return
	}

	var b strings.Builder
	if r.StatusMsg == "Accepted" {
		b.WriteString("✅ Accepted\n")
	} else {
		b.WriteString(fmt.Sprintf("❌ %s\n", r.StatusMsg))
	}

	if len(r.CodeAnswer) > 0 {
		b.WriteString(fmt.Sprintf("\nOutput:   %s\n", strings.Join(r.CodeAnswer, ", ")))
	}
	if len(r.ExpectedAns) > 0 {
		b.WriteString(fmt.Sprintf("Expected: %s\n", strings.Join(r.ExpectedAns, ", ")))
	}
	if r.CompileError != "" {
		b.WriteString(fmt.Sprintf("\nCompile Error:\n  %s\n", r.CompileError))
	}
	if r.RuntimeError != "" {
		b.WriteString(fmt.Sprintf("\nRuntime Error:\n  %s\n", r.RuntimeError))
	}

	d.content = strings.TrimRight(b.String(), "\n")
}

// Update handles keyboard input for the result dialog.
func (d *ResultDialog) Update(msg tea.Msg) (*ResultDialog, tea.Cmd) {
	if d.loading {
		return d, nil
	}
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "esc", "enter", "q":
			return d, func() tea.Msg { return ResultDismissMsg{} }
		}
	}
	return d, nil
}

// View renders the result dialog.
func (d *ResultDialog) View() string {
	body := resultTitleStyle.Render(d.title) + "\n" +
		resultContentStyle.Render(d.content)
	if !d.loading {
		body += "\n" + resultHintStyle.Render("Press esc/enter to close")
	}
	return resultDialogStyle.Render(body)
}
