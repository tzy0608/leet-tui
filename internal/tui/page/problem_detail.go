package page

import (
	"fmt"
	"os"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tzy0608/leet-tui/internal/editor"
	"github.com/tzy0608/leet-tui/internal/leetcode"
	"github.com/tzy0608/leet-tui/internal/tui/components/dialog"
	"github.com/tzy0608/leet-tui/internal/tui/styles"
	"github.com/tzy0608/leet-tui/internal/tui/theme"
)

// ProblemDetailLoadedMsg is sent when problem details are fetched.
type ProblemDetailLoadedMsg struct {
	Problem *leetcode.Problem
}

// AcceptedCodeLoadedMsg is sent when the latest accepted code is fetched asynchronously.
type AcceptedCodeLoadedMsg struct {
	Slug string
	Code string
}

// SubmitResultMsg is sent when a submission result is available.
type SubmitResultMsg struct {
	Result *leetcode.SubmitResult
	Err    error
}

// RunResultMsg is sent when a run result is available.
type RunResultMsg struct {
	Result *leetcode.RunResult
	Err    error
}

// RunTestRequestMsg requests running code against sample test cases.
type RunTestRequestMsg struct {
	TitleSlug  string
	QuestionID int
	Lang       string
	Code       string
	TestInput  string
}

// SubmitRequestMsg requests submitting code for evaluation.
type SubmitRequestMsg struct {
	TitleSlug  string
	QuestionID int
	Lang       string
	Code       string
}

// PanelFocus identifies which panel is focused.
type PanelFocus int

const (
	PanelDescription PanelFocus = iota
	PanelCode
)

// ProblemDetail is the problem detail page with description and code.
type ProblemDetail struct {
	problem    *leetcode.Problem
	descView   viewport.Model
	codeView   viewport.Model
	panelFocus PanelFocus
	focused    bool
	width      int
	height     int
	lang       string
	code       string
	editorCmd  string
	loading    bool
	statusMsg  string
	resultDlg  *dialog.ResultDialog
}

// NewProblemDetail creates a new problem detail page.
func NewProblemDetail(editorCmd, lang string) *ProblemDetail {
	return &ProblemDetail{
		editorCmd:  editorCmd,
		lang:       lang,
		panelFocus: PanelDescription,
		descView:   viewport.New(0, 0),
		codeView:   viewport.New(0, 0),
	}
}

// SetProblem updates the displayed problem.
func (p *ProblemDetail) SetProblem(problem *leetcode.Problem) {
	p.problem = problem

	// Try loading previously edited code from temp file first
	tmpPath := editor.TempFilePath(p.lang, problem.TitleSlug)
	if data, err := os.ReadFile(tmpPath); err == nil && len(data) > 0 {
		p.code = string(data)
	} else if problem.LastAcceptedCode != "" {
		p.code = problem.LastAcceptedCode
	} else {
		slug := leetcode.LangSlug(p.lang)
		p.code = problem.CodeSnippets[slug]
		if p.code == "" {
			// Try original value as fallback
			p.code = problem.CodeSnippets[p.lang]
		}
		if p.code == "" {
			// Last resort: use first available snippet
			for _, snippet := range problem.CodeSnippets {
				p.code = snippet
				break
			}
		}
	}
	p.refreshViews()
}

func (p *ProblemDetail) refreshViews() {
	p.refreshDescView()
	p.refreshCodeView()
}

func (p *ProblemDetail) refreshDescView() {
	if p.problem == nil {
		return
	}

	// Content width = viewport width minus border (2 chars)
	contentW := p.descView.Width - 2
	if contentW < 20 {
		contentW = 80 // fallback before resize
	}

	// Render description as markdown with correct width
	descMD := htmlToMarkdown(p.problem.Content)
	p.descView.SetContent(renderMarkdown(descMD, contentW))
}

func (p *ProblemDetail) refreshCodeView() {
	// Render code with line numbers, truncated to panel width
	p.codeView.SetContent(formatCode(p.code, p.lang, p.codeView.Width))
}

func (p *ProblemDetail) resize() {
	// Each panel: content width + 2 (border left+right) = visual width
	// Two panels total visual width must be <= p.width
	leftW := p.width/2 - 2
	rightW := p.width - p.width/2 - 2
	// header(1) + blank line(1) + border top/bottom(2) + statusbar line(1) = 5
	contentH := p.height - 5

	p.descView.Width = leftW
	p.descView.Height = contentH

	p.codeView.Width = rightW
	p.codeView.Height = contentH - 1 // code header occupies 1 line

	// Re-render glamour with new width
	p.refreshViews()
}

func (p *ProblemDetail) Init() tea.Cmd { return nil }

func (p *ProblemDetail) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case ProblemDetailLoadedMsg:
		p.loading = false
		p.SetProblem(msg.Problem)

	case AcceptedCodeLoadedMsg:
		if p.problem != nil && p.problem.TitleSlug == msg.Slug {
			p.code = msg.Code
			p.refreshCodeView()
		}

	case editor.OpenMsg:
		if msg.Err == nil && msg.Code != "" {
			p.code = msg.Code
			p.refreshCodeView()
		}
		p.statusMsg = ""

	case SubmitResultMsg:
		p.loading = false
		if p.resultDlg != nil {
			p.resultDlg.SetSubmitResult(msg.Result, msg.Err)
		}

	case RunResultMsg:
		p.loading = false
		if p.resultDlg != nil {
			p.resultDlg.SetRunResult(msg.Result, msg.Err)
		}

	case dialog.ResultDismissMsg:
		p.resultDlg = nil
		return p, nil

	case tea.KeyMsg:
		if !p.focused {
			return p, nil
		}
		// Route keys to result dialog when open
		if p.resultDlg != nil {
			var cmd tea.Cmd
			p.resultDlg, cmd = p.resultDlg.Update(msg)
			cmds = append(cmds, cmd)
			return p, tea.Batch(cmds...)
		}
		switch msg.String() {
		case "tab":
			if p.panelFocus == PanelDescription {
				p.panelFocus = PanelCode
			} else {
				p.panelFocus = PanelDescription
			}
		case "y":
			if p.code != "" {
				if err := clipboard.WriteAll(p.code); err != nil {
					p.statusMsg = "Copy failed: " + err.Error()
				} else {
					p.statusMsg = "Code copied to clipboard"
				}
			}
		case "e":
			if p.problem != nil {
				p.statusMsg = "Opening editor..."
				return p, editor.Open(p.editorCmd, p.lang, p.code, p.problem.TitleSlug)
			}
		case "r":
			if p.problem != nil && p.code != "" {
				p.loading = true
				p.resultDlg = dialog.NewResultDialog("Run Test")
				return p, func() tea.Msg {
					return RunTestRequestMsg{
						TitleSlug:  p.problem.TitleSlug,
						QuestionID: p.problem.ID,
						Lang:       leetcode.LangSlug(p.lang),
						Code:       p.code,
						TestInput:  p.problem.SampleTestCase,
					}
				}
			}
		case "s":
			if p.problem != nil && p.code != "" {
				p.loading = true
				p.resultDlg = dialog.NewResultDialog("Submit")
				return p, func() tea.Msg {
					return SubmitRequestMsg{
						TitleSlug:  p.problem.TitleSlug,
						QuestionID: p.problem.ID,
						Lang:       leetcode.LangSlug(p.lang),
						Code:       p.code,
					}
				}
			}
		case "esc", "backspace":
			return p, NavigateTo(ProblemsPage)
		}
	}

	// Route scroll to focused panel
	var cmd tea.Cmd
	if p.panelFocus == PanelDescription {
		p.descView, cmd = p.descView.Update(msg)
	} else {
		p.codeView, cmd = p.codeView.Update(msg)
	}
	cmds = append(cmds, cmd)

	return p, tea.Batch(cmds...)
}

func (p *ProblemDetail) Focus() tea.Cmd { p.focused = true; return nil }
func (p *ProblemDetail) Blur() tea.Cmd  { p.focused = false; return nil }
func (p *ProblemDetail) IsFocused() bool { return p.focused }

func (p *ProblemDetail) SetSize(w, h int) tea.Cmd {
	p.width = w
	p.height = h
	p.resize()
	return nil
}

func (p *ProblemDetail) GetSize() (int, int) { return p.width, p.height }

func (p *ProblemDetail) BindingKeys() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "switch panel")),
		key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit in $EDITOR")),
		key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "run test")),
		key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "submit")),
		key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "copy code")),
		key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	}
}

func (p *ProblemDetail) View() string {
	if p.width == 0 {
		return "Loading..."
	}

	if p.problem == nil {
		return lipgloss.Place(p.width, p.height, lipgloss.Center, lipgloss.Center, "Loading problem...")
	}

	leftW := p.width/2 - 2
	rightW := p.width - p.width/2 - 2
	contentH := p.height - 5

	// Header
	diff := styles.DifficultyStyle(p.problem.Difficulty).Bold(true).Render(p.problem.Difficulty)
	header := lipgloss.NewStyle().Bold(true).Foreground(theme.Current.Primary).
		Render(fmt.Sprintf("#%s %s", p.problem.FrontendID, p.problem.Title)) + " " + diff

	// Panel border styles (no padding — viewport handles content)
	descBorder := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current.Border)
	if p.panelFocus == PanelDescription {
		descBorder = descBorder.BorderForeground(theme.Current.Primary)
	}

	codeBorder := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current.Border)
	if p.panelFocus == PanelCode {
		codeBorder = codeBorder.BorderForeground(theme.Current.Primary)
	}

	// Left: description
	descPanel := descBorder.Width(leftW).Height(contentH).
		Render(p.descView.View())

	// Right: code
	langLabel := lipgloss.NewStyle().
		Foreground(theme.Current.Accent).Bold(true).
		Render("[" + p.lang + "]")
	codeHeader := lipgloss.NewStyle().Bold(true).
		Foreground(theme.Current.Text).
		Render("Code ") + langLabel

	codePanel := codeBorder.Width(rightW).Height(contentH).
		Render(codeHeader + "\n" + p.codeView.View())

	body := lipgloss.JoinHorizontal(lipgloss.Top, descPanel, codePanel)

	// Status bar
	status := p.statusMsg
	if status == "" {
		status = "Tab:switch  e:edit  r:run  s:submit  y:copy  esc:back"
	}
	statusBar := lipgloss.NewStyle().
		Foreground(theme.Current.TextMuted).
		Render("  " + status)

	page := header + "\n\n" + body + "\n" + statusBar

	if p.resultDlg != nil {
		dlgView := p.resultDlg.View()
		return lipgloss.Place(p.width, p.height, lipgloss.Center, lipgloss.Center, dlgView)
	}

	return page
}

// htmlToMarkdown converts HTML content to markdown (simplified).
func htmlToMarkdown(html string) string {
	// Minimal HTML to Markdown conversion for LeetCode content
	md := html
	md = strings.ReplaceAll(md, "<p>", "")
	md = strings.ReplaceAll(md, "</p>", "\n\n")
	md = strings.ReplaceAll(md, "<br>", "\n")
	md = strings.ReplaceAll(md, "<br/>", "\n")
	md = strings.ReplaceAll(md, "<strong>", "**")
	md = strings.ReplaceAll(md, "</strong>", "**")
	md = strings.ReplaceAll(md, "<em>", "_")
	md = strings.ReplaceAll(md, "</em>", "_")
	md = strings.ReplaceAll(md, "<code>", "`")
	md = strings.ReplaceAll(md, "</code>", "`")
	md = strings.ReplaceAll(md, "<pre>", "```\n")
	md = strings.ReplaceAll(md, "</pre>", "\n```")
	md = strings.ReplaceAll(md, "<ul>", "")
	md = strings.ReplaceAll(md, "</ul>", "")
	md = strings.ReplaceAll(md, "<li>", "- ")
	md = strings.ReplaceAll(md, "</li>", "\n")
	md = strings.ReplaceAll(md, "<ol>", "")
	md = strings.ReplaceAll(md, "</ol>", "")
	md = strings.ReplaceAll(md, "<sup>", "^(")
	md = strings.ReplaceAll(md, "</sup>", ")")
	md = strings.ReplaceAll(md, "<sub>", "_(")
	md = strings.ReplaceAll(md, "</sub>", ")")

	// Remove remaining HTML tags (but not bare < from entities like &lt;)
	var result strings.Builder
	i := 0
	for i < len(md) {
		if md[i] == '<' {
			// Look for closing '>'
			end := strings.Index(md[i:], ">")
			if end < 0 {
				// No closing '>', keep the '<' literally
				result.WriteByte(md[i])
				i++
				continue
			}
			tagContent := md[i+1 : i+end]
			// Check if this looks like an HTML tag (starts with letter, /, or !)
			if len(tagContent) > 0 && (tagContent[0] == '/' ||
				(tagContent[0] >= 'a' && tagContent[0] <= 'z') ||
				(tagContent[0] >= 'A' && tagContent[0] <= 'Z') ||
				tagContent[0] == '!') {
				// Skip the tag
				i += end + 1
			} else {
				// Not an HTML tag, keep the '<'
				result.WriteByte(md[i])
				i++
			}
		} else {
			result.WriteByte(md[i])
			i++
		}
	}
	md = result.String()

	// Decode HTML entities last (after tag removal, so decoded < won't be re-processed)
	md = strings.ReplaceAll(md, "&lt;", "<")
	md = strings.ReplaceAll(md, "&gt;", ">")
	md = strings.ReplaceAll(md, "&amp;", "&")
	md = strings.ReplaceAll(md, "&quot;", "\"")
	md = strings.ReplaceAll(md, "&#39;", "'")
	md = strings.ReplaceAll(md, "&nbsp;", " ")

	// Collapse excessive blank lines (e.g. </p>\n\n<p> produces 4+ newlines)
	for strings.Contains(md, "\n\n\n") {
		md = strings.ReplaceAll(md, "\n\n\n", "\n\n")
	}
	md = strings.TrimSpace(md)

	return md
}

// formatCode adds line numbers to code and truncates lines to maxWidth.
func formatCode(code, lang string, maxWidth int) string {
	lines := strings.Split(code, "\n")
	var result []string
	numWidth := len(fmt.Sprintf("%d", len(lines)))
	numStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#9CA3AF", Dark: "#585B70"})
	sepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#CCD0DA", Dark: "#313244"})

	// prefix width: numWidth + " │ " (3 visible chars)
	prefixW := numWidth + 3
	codeW := maxWidth - prefixW
	if codeW < 10 {
		codeW = 10
	}

	for i, line := range lines {
		line = strings.ReplaceAll(line, "\t", "    ")
		num := numStyle.Render(fmt.Sprintf("%*d", numWidth, i+1))
		sep := sepStyle.Render(" │ ")
		if len(line) > codeW {
			line = line[:codeW-1] + "…"
		}
		result = append(result, num+sep+line)
	}
	return strings.Join(result, "\n")
}
