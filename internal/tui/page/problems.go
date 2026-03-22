package page

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leet-tui/leet-tui/internal/leetcode"
	"github.com/leet-tui/leet-tui/internal/tui/components/problemlist"
	"github.com/leet-tui/leet-tui/internal/tui/components/searchbar"
	"github.com/leet-tui/leet-tui/internal/tui/components/topicpanel"
	"github.com/leet-tui/leet-tui/internal/tui/styles"
	"github.com/leet-tui/leet-tui/internal/tui/theme"
)

// ProblemsLoadedMsg is sent when problems are fetched from DB.
type ProblemsLoadedMsg struct {
	Items []leetcode.ProblemListItem
	Tags  []topicpanel.TagItem
	Total int
}

// ProblemStatusUpdatedMsg carries AC statuses fetched from the API.
type ProblemStatusUpdatedMsg struct {
	Statuses map[string]string // titleSlug -> status ("ac", "notac", "")
}

// FocusArea identifies which sub-component has focus.
type FocusArea int

const (
	FocusTopic FocusArea = iota
	FocusList
	FocusSearch

	topicPanelWidth = 18
)

// Problems is the problem browsing page.
type Problems struct {
	topicPanel  *topicpanel.Panel
	problemList *problemlist.List
	searchBar   *searchbar.Bar
	focusArea   FocusArea
	focused     bool
	width       int
	height      int
}

// NewProblems creates a new problems page.
func NewProblems() *Problems {
	return &Problems{
		topicPanel:  topicpanel.New(),
		problemList: problemlist.New(),
		searchBar:   searchbar.New(),
		focusArea:   FocusList,
	}
}

// SetData populates the page with loaded problems.
func (p *Problems) SetData(msg ProblemsLoadedMsg) {
	p.topicPanel.SetTags(msg.Tags, msg.Total)
	p.problemList.SetItems(msg.Items)
	p.resize()
}

func (p *Problems) resize() {
	topicW := topicPanelWidth
	listW := p.width - topicW - 4 // -4: two borders × 2 chars each
	listH := p.height - 5        // search bar + header

	p.topicPanel.SetSize(topicW, p.height-2)
	p.searchBar.SetWidth(listW)
	p.problemList.SetSize(listW, listH)
}

func (p *Problems) Init() tea.Cmd { return nil }

func (p *Problems) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case ProblemsLoadedMsg:
		p.SetData(msg)

	case ProblemStatusUpdatedMsg:
		p.problemList.UpdateStatuses(msg.Statuses)

	case problemlist.SelectMsg:
		return p, NavigateTo(ProblemDetailPage, map[string]any{
			"slug": msg.Problem.TitleSlug,
			"id":   msg.Problem.ID,
		})

	case topicpanel.SelectMsg:
		// Filter problem list by selected tag
		p.focusArea = FocusList
		p.topicPanel.Blur()
		p.problemList.Focus()
		p.problemList.FilterByTag(msg.Tag)

	case searchbar.SearchMsg:
		p.problemList.Filter(msg.Query)

	case tea.KeyMsg:
		if !p.focused {
			return p, nil
		}
		switch msg.String() {
		case "/":
			p.focusArea = FocusSearch
			p.problemList.Blur()
			p.topicPanel.Blur()
			cmds = append(cmds, p.searchBar.Focus())
			return p, tea.Batch(cmds...)
		case "esc":
			if p.focusArea == FocusSearch {
				// Exit search bar, keep filter results
				p.searchBar.Blur()
				p.focusArea = FocusList
				p.problemList.Focus()
			} else if p.focusArea == FocusList && p.searchBar.Value() != "" {
				// Second Esc: clear search query and reset filter
				p.searchBar.Clear()
				p.problemList.Filter("")
			}
		case "tab":
			p.cycleFocus()
		}
	}

	// Route message to focused component
	switch p.focusArea {
	case FocusList:
		var cmd tea.Cmd
		p.problemList, cmd = p.problemList.Update(msg)
		cmds = append(cmds, cmd)
	case FocusTopic:
		var cmd tea.Cmd
		p.topicPanel, cmd = p.topicPanel.Update(msg)
		cmds = append(cmds, cmd)
	case FocusSearch:
		var cmd tea.Cmd
		p.searchBar, cmd = p.searchBar.Update(msg)
		cmds = append(cmds, cmd)
	}

	return p, tea.Batch(cmds...)
}

func (p *Problems) cycleFocus() {
	switch p.focusArea {
	case FocusTopic:
		p.topicPanel.Blur()
		p.focusArea = FocusList
		p.problemList.Focus()
	case FocusList:
		p.problemList.Blur()
		p.focusArea = FocusTopic
		p.topicPanel.Focus()
	}
}

func (p *Problems) Focus() tea.Cmd {
	p.focused = true
	p.problemList.Focus()
	return nil
}

func (p *Problems) Blur() tea.Cmd {
	p.focused = false
	p.problemList.Blur()
	p.topicPanel.Blur()
	p.searchBar.Blur()
	return nil
}

func (p *Problems) IsFocused() bool { return p.focused }

func (p *Problems) IsInputFocused() bool {
	return p.focused && p.focusArea == FocusSearch
}

func (p *Problems) SetSize(w, h int) tea.Cmd {
	p.width = w
	p.height = h
	p.resize()
	return nil
}

func (p *Problems) GetSize() (int, int) { return p.width, p.height }

func (p *Problems) BindingKeys() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "switch focus")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open problem")),
		key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "clear search")),
	}
}

func (p *Problems) View() string {
	if p.width == 0 {
		return "Loading..."
	}

	topicW := topicPanelWidth
	listW := p.width - topicW - 4 // -4: two borders × 2 chars each

	topicBorder := styles.Border
	listBorder := styles.Border
	if p.focusArea == FocusTopic {
		topicBorder = styles.FocusedBorder
	} else if p.focusArea == FocusList || p.focusArea == FocusSearch {
		listBorder = styles.FocusedBorder
	}

	topicView := topicBorder.Width(topicW).Height(p.height - 2).Render(p.topicPanel.View())

	searchView := p.searchBar.View()
	listView := p.problemList.View()
	rightContent := searchView + "\n" + listView
	rightView := listBorder.Width(listW).Height(p.height - 2).Render(rightContent)

	hint := lipgloss.NewStyle().
		Foreground(theme.Current.TextMuted).
		Render("  / Search  Tab Switch  Enter Open")

	content := lipgloss.JoinHorizontal(lipgloss.Top, topicView, rightView)
	return content + "\n" + hint
}
