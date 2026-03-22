package topicpanel

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tzy0608/leet-tui/internal/tui/styles"
	"github.com/tzy0608/leet-tui/internal/tui/theme"
)

// SelectMsg is sent when a topic is selected.
type SelectMsg struct {
	Tag string // empty = all
}

// Panel is the topic/tag sidebar.
type Panel struct {
	tags    []TagItem
	cursor  int
	focused bool
	width   int
	height  int
}

// TagItem is a tag with a count.
type TagItem struct {
	Name  string
	Count int
}

var (
	selectedStyle = lipgloss.NewStyle().
		Background(lipgloss.AdaptiveColor{Light: "#DDD6FE", Dark: "#3B2D63"}).
		Foreground(theme.Current.Primary).
		Bold(true)

	normalStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#4C4F69", Dark: "#CDD6F4"})

	countStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#9CA3AF", Dark: "#585B70"})

	topicTitleStyle = styles.PanelHeader
)

// New creates a new topic panel.
func New() *Panel {
	return &Panel{}
}

// SetTags updates the tag list.
func (p *Panel) SetTags(tags []TagItem, totalCount int) {
	p.tags = append([]TagItem{{Name: "All", Count: totalCount}}, tags...)
	p.cursor = 0
}

// Focus focuses the panel.
func (p *Panel) Focus() tea.Cmd {
	p.focused = true
	return nil
}

// Blur unfocuses the panel.
func (p *Panel) Blur() tea.Cmd {
	p.focused = false
	return nil
}

// SetSize updates panel dimensions.
func (p *Panel) SetSize(w, h int) tea.Cmd {
	p.width = w
	p.height = h
	return nil
}

// Update handles messages.
func (p *Panel) Update(msg tea.Msg) (*Panel, tea.Cmd) {
	if !p.focused {
		return p, nil
	}
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "j", "down":
			if p.cursor < len(p.tags)-1 {
				p.cursor++
			}
		case "k", "up":
			if p.cursor > 0 {
				p.cursor--
			}
		case "enter":
			tag := ""
			if p.cursor > 0 && p.cursor < len(p.tags) {
				tag = p.tags[p.cursor].Name
			}
			return p, func() tea.Msg { return SelectMsg{Tag: tag} }
		}
	}
	return p, nil
}

// View renders the topic panel.
func (p *Panel) View() string {
	header := topicTitleStyle.Width(p.width - 2).Render("Topics")
	var rows []string
	rows = append(rows, header)

	maxVisible := p.height - 3
	for i, tag := range p.tags {
		if i >= maxVisible {
			break
		}

		label := tag.Name
		if len(label) > p.width-7 {
			label = label[:p.width-10] + ".."
		}

		count := countStyle.Render(fmt.Sprintf("%d", tag.Count))
		// Right-align count
		labelW := p.width - 8
		line := fmt.Sprintf("%-*s", labelW, label) + " " + count

		if i == p.cursor {
			indicator := "▌ "
			if !p.focused {
				indicator = "  "
			}
			rows = append(rows, selectedStyle.Width(p.width-2).Padding(0, 1).Render(indicator+line))
		} else {
			rowStyle := styles.RowEven
			if i%2 == 1 {
				rowStyle = styles.RowOdd
			}
			rows = append(rows, rowStyle.Copy().
				Foreground(lipgloss.AdaptiveColor{Light: "#4C4F69", Dark: "#CDD6F4"}).
				Width(p.width-2).Render("  "+line))
		}
	}

	// Pad remaining
	for len(rows) < p.height-1 {
		rows = append(rows, strings.Repeat(" ", p.width))
	}

	return strings.Join(rows, "\n")
}

// SelectedTag returns the currently selected tag name.
func (p *Panel) SelectedTag() string {
	if p.cursor > 0 && p.cursor < len(p.tags) {
		return p.tags[p.cursor].Name
	}
	return ""
}
