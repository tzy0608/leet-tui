package problemlist

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leet-tui/leet-tui/internal/leetcode"
	"github.com/leet-tui/leet-tui/internal/tui/styles"
	"github.com/leet-tui/leet-tui/internal/tui/theme"
)

// SelectMsg is sent when a problem is selected.
type SelectMsg struct {
	Problem leetcode.ProblemListItem
}

// List is a scrollable, filterable problem list.
type List struct {
	items    []leetcode.ProblemListItem
	filtered []leetcode.ProblemListItem
	cursor    int
	offset    int
	width     int
	height    int
	focused   bool
	query     string
	tagFilter string
}

var (
	headerStyle = styles.PanelHeader

	selectedRowStyle = lipgloss.NewStyle().
		Background(lipgloss.AdaptiveColor{Light: "#DDD6FE", Dark: "#3B2D63"}).
		Foreground(theme.Current.Primary)

	normalRowStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#4C4F69", Dark: "#CDD6F4"})

	acceptedStyle = lipgloss.NewStyle().Foreground(theme.Current.Success)
	mutedStyle    = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#9CA3AF", Dark: "#585B70"})

	countSepStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(lipgloss.AdaptiveColor{Light: "#E6E9EF", Dark: "#313244"})
)

// New creates a new problem list.
func New() *List {
	return &List{}
}

// SetItems replaces the problem list.
func (l *List) SetItems(items []leetcode.ProblemListItem) {
	l.items = items
	l.applyFilter()
	l.cursor = 0
	l.offset = 0
}

// Filter applies a search query.
func (l *List) Filter(query string) {
	l.query = strings.ToLower(query)
	l.applyFilter()
	l.cursor = 0
	l.offset = 0
}

// FilterByTag filters problems by topic tag. Empty tag shows all.
func (l *List) FilterByTag(tag string) {
	l.tagFilter = tag
	l.applyFilter()
	l.cursor = 0
	l.offset = 0
}

func (l *List) applyFilter() {
	l.filtered = l.items

	// Apply tag filter
	if l.tagFilter != "" {
		var tagFiltered []leetcode.ProblemListItem
		for _, item := range l.filtered {
			for _, t := range item.TopicTags {
				if t == l.tagFilter {
					tagFiltered = append(tagFiltered, item)
					break
				}
			}
		}
		l.filtered = tagFiltered
	}

	// Apply text search filter
	if l.query != "" {
		var textFiltered []leetcode.ProblemListItem
		for _, item := range l.filtered {
			if strings.Contains(strings.ToLower(item.Title), l.query) ||
				strings.Contains(item.FrontendID, l.query) {
				textFiltered = append(textFiltered, item)
			}
		}
		l.filtered = textFiltered
	}
}

// Focus focuses the list.
func (l *List) Focus() tea.Cmd {
	l.focused = true
	return nil
}

// Blur unfocuses the list.
func (l *List) Blur() tea.Cmd {
	l.focused = false
	return nil
}

// SetSize updates the list dimensions.
func (l *List) SetSize(w, h int) tea.Cmd {
	l.width = w
	l.height = h
	return nil
}

// BindingKeys returns key bindings.
func (l *List) BindingKeys() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/↓", "move down")),
		key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k/↑", "move up")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open problem")),
		key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "go to top")),
		key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "go to bottom")),
	}
}

// Update handles messages.
func (l *List) Update(msg tea.Msg) (*List, tea.Cmd) {
	if !l.focused {
		return l, nil
	}

	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "j", "down":
			if l.cursor < len(l.filtered)-1 {
				l.cursor++
				l.adjustOffset()
			}
		case "k", "up":
			if l.cursor > 0 {
				l.cursor--
				l.adjustOffset()
			}
		case "g":
			l.cursor = 0
			l.offset = 0
		case "G":
			l.cursor = len(l.filtered) - 1
			l.adjustOffset()
		case "enter":
			if l.cursor < len(l.filtered) {
				item := l.filtered[l.cursor]
				return l, func() tea.Msg { return SelectMsg{Problem: item} }
			}
		}
	}
	return l, nil
}

func (l *List) adjustOffset() {
	visibleRows := l.visibleRows()
	if l.cursor < l.offset {
		l.offset = l.cursor
	} else if l.cursor >= l.offset+visibleRows {
		l.offset = l.cursor - visibleRows + 1
	}
}

func (l *List) visibleRows() int {
	// header(1) + rows
	rows := l.height - 2 // header + padding
	if rows < 1 {
		rows = 1
	}
	return rows
}

// View renders the problem list.
func (l *List) View() string {
	if l.width == 0 {
		return ""
	}

	// Column widths
	numW := 5
	diffW := 7
	statusW := 6
	titleW := l.width - numW - diffW - statusW - 3

	// Header
	header := headerStyle.Width(l.width).Render(
		fmt.Sprintf("%-*s %-*s %-*s %s",
			numW, "#",
			titleW, "Title",
			diffW, "Diff",
			"Status",
		),
	)

	// Rows
	visible := l.visibleRows()
	end := l.offset + visible
	if end > len(l.filtered) {
		end = len(l.filtered)
	}

	var rows []string
	for i := l.offset; i < end; i++ {
		item := l.filtered[i]
		row := l.renderRow(item, i == l.cursor, i, numW, titleW, diffW, statusW)
		rows = append(rows, row)
	}

	// Pad empty rows
	for len(rows) < visible {
		rows = append(rows, strings.Repeat(" ", l.width))
	}

	count := countSepStyle.Width(l.width).Render(
		mutedStyle.Render(fmt.Sprintf(" %d/%d problems", len(l.filtered), len(l.items))),
	)
	return header + "\n" + strings.Join(rows, "\n") + "\n" + count
}

func (l *List) renderRow(item leetcode.ProblemListItem, selected bool, index int, numW, titleW, diffW, statusW int) string {
	diffStr := styles.DifficultyStyle(item.Difficulty).Render(
		fmt.Sprintf("%-*s", diffW, item.Difficulty),
	)

	status := mutedStyle.Render("·")
	if item.Status == "ac" {
		status = acceptedStyle.Render("✓")
	}

	title := item.Title
	titleRunes := []rune(title)
	if len(titleRunes) > titleW {
		title = string(titleRunes[:titleW-2]) + ".."
	}

	row := fmt.Sprintf("%-*s %-*s %s %s",
		numW, item.FrontendID,
		titleW, title,
		diffStr,
		status,
	)

	if selected {
		indicator := "▌ "
		if !l.focused {
			indicator = "  "
		}
		return selectedRowStyle.
			Width(l.width).
			Padding(0, 1).
			Render(indicator + row)
	}

	// Alternating row background
	rowStyle := styles.RowEven
	if index%2 == 1 {
		rowStyle = styles.RowOdd
	}
	return rowStyle.Copy().
		Foreground(lipgloss.AdaptiveColor{Light: "#4C4F69", Dark: "#CDD6F4"}).
		Width(l.width).
		Render("  " + row)
}

// UpdateStatuses updates the AC status of items by titleSlug.
func (l *List) UpdateStatuses(statuses map[string]string) {
	for i := range l.items {
		if s, ok := statuses[l.items[i].TitleSlug]; ok {
			l.items[i].Status = s
		}
	}
	l.applyFilter()
}

// SelectedItem returns the currently highlighted problem.
func (l *List) SelectedItem() *leetcode.ProblemListItem {
	if l.cursor < len(l.filtered) {
		item := l.filtered[l.cursor]
		return &item
	}
	return nil
}
