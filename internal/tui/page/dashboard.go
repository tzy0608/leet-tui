package page

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tzy0608/leet-tui/internal/study"
	"github.com/tzy0608/leet-tui/internal/tui/components/stats"
	"github.com/tzy0608/leet-tui/internal/tui/styles"
	"github.com/tzy0608/leet-tui/internal/tui/theme"
)

// DashboardLoadedMsg carries data needed to render the dashboard.
type DashboardLoadedMsg struct {
	Queue *study.DailyQueue
	Stats *study.Stats
}

// DashboardStatsMsg carries refreshed stats for the dashboard.
type DashboardStatsMsg struct {
	Stats *study.Stats
}

// Dashboard is the home page showing the daily queue and stats.
type Dashboard struct {
	queue      *study.DailyQueue
	stats      *study.Stats
	queueItems []study.QueueItem // flattened list for cursor navigation
	cursor     int
	width      int
	height     int
	focused    bool
}

var (
	dashPanelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current.Border).
		Padding(1, 2)

	dashTitleStyle = styles.PanelHeader

	dashSectionStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Current.Secondary).
		MarginTop(1)

	queueItemStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#4C4F69", Dark: "#CDD6F4"})

	queueSelectedStyle = lipgloss.NewStyle().
		Background(lipgloss.AdaptiveColor{Light: "#DDD6FE", Dark: "#3B2D63"}).
		Foreground(theme.Current.Primary).
		Padding(0, 1)

	reviewBadge = styles.NewBadge(
		lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#1E1E2E"},
		theme.Current.Warning,
	)

	newBadge = styles.NewBadge(
		lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#1E1E2E"},
		theme.Current.Secondary,
	)
)

// NewDashboard creates a new dashboard page.
func NewDashboard() *Dashboard {
	return &Dashboard{}
}

// SetData updates the dashboard with loaded data.
func (d *Dashboard) SetData(msg DashboardLoadedMsg) {
	d.queue = msg.Queue
	d.stats = msg.Stats

	// Build flattened queue items list for cursor navigation
	d.queueItems = nil
	if d.queue != nil {
		d.queueItems = append(d.queueItems, d.queue.ReviewItems...)
		d.queueItems = append(d.queueItems, d.queue.NewItems...)
	}
	if d.cursor >= len(d.queueItems) {
		d.cursor = 0
	}
}

func (d *Dashboard) Init() tea.Cmd { return nil }

func (d *Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case DashboardLoadedMsg:
		d.SetData(msg)
	case DashboardStatsMsg:
		d.stats = msg.Stats
	case tea.KeyMsg:
		if !d.focused {
			return d, nil
		}
		switch msg.String() {
		case "j", "down":
			if len(d.queueItems) > 0 && d.cursor < len(d.queueItems)-1 {
				d.cursor++
			}
		case "k", "up":
			if d.cursor > 0 {
				d.cursor--
			}
		case "enter":
			if d.cursor < len(d.queueItems) {
				item := d.queueItems[d.cursor]
				return d, NavigateTo(ProblemDetailPage, map[string]any{
					"slug": item.TitleSlug,
				})
			}
		case "r":
			return d, NavigateTo(ReviewPage)
		case "p":
			return d, NavigateTo(ProblemsPage)
		}
	}
	return d, nil
}

func (d *Dashboard) Focus() tea.Cmd { d.focused = true; return nil }
func (d *Dashboard) Blur() tea.Cmd  { d.focused = false; return nil }
func (d *Dashboard) IsFocused() bool { return d.focused }

func (d *Dashboard) SetSize(w, h int) tea.Cmd {
	d.width = w
	d.height = h
	return nil
}

func (d *Dashboard) GetSize() (int, int) { return d.width, d.height }

func (d *Dashboard) BindingKeys() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/↓", "next")),
		key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k/↑", "prev")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open problem")),
		key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "start review")),
		key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "browse problems")),
	}
}

func (d *Dashboard) View() string {
	if d.width == 0 {
		return "Loading..."
	}

	// Welcome header
	welcome := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Current.Primary).
		Render("LeetCode TUI")

	panelH := d.height - 4 // title 1 line + blank 1 line + outer border top/bottom 2 lines
	leftW := (d.width - 2) / 2
	rightW := d.width - 2 - leftW - 1

	leftContent := d.renderQueue(leftW - 4)
	rightContent := d.renderStats(rightW - 4)

	leftCol := lipgloss.NewStyle().Width(leftW).Height(panelH).Padding(1, 2).Render(leftContent)
	rightCol := lipgloss.NewStyle().Width(rightW).Height(panelH).Padding(1, 2).Render(rightContent)
	divider := styles.VerticalDivider(panelH)

	inner := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, divider, rightCol)
	outer := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current.Border).
		Render(inner)

	return welcome + "\n" + outer
}

func (d *Dashboard) renderQueue(w int) string {
	var lines []string
	lines = append(lines, dashTitleStyle.Width(w).Render("Today's Queue"))

	if d.queue == nil {
		lines = append(lines, "Loading...")
		return strings.Join(lines, "\n")
	}

	if d.queue.Total() == 0 {
		lines = append(lines, "")
		lines = append(lines, lipgloss.NewStyle().Foreground(theme.Current.Success).Render("  ✓ All caught up!"))
		lines = append(lines, "")
		lines = append(lines, lipgloss.NewStyle().Foreground(theme.Current.TextMuted).Render("  Quick actions:"))
		lines = append(lines, lipgloss.NewStyle().Foreground(theme.Current.TextMuted).Render("    p  Browse problems"))
		lines = append(lines, lipgloss.NewStyle().Foreground(theme.Current.TextMuted).Render("    4  Study plans"))
		lines = append(lines, lipgloss.NewStyle().Foreground(theme.Current.TextMuted).Render("    ?  Help"))
	} else {
		// Track global item index for cursor highlighting
		itemIdx := 0

		if len(d.queue.ReviewItems) > 0 {
			lines = append(lines, dashSectionStyle.Render(reviewBadge.Render("REVIEW")))
			lines = append(lines, "")
			for _, item := range d.queue.ReviewItems {
				lines = append(lines, d.renderQueueItem(item, itemIdx, w))
				itemIdx++
			}
		} else {
			lines = append(lines, queueItemStyle.Render("  ✓ No reviews due today!"))
		}

		lines = append(lines, "")

		if len(d.queue.NewItems) > 0 {
			lines = append(lines, dashSectionStyle.Render(newBadge.Render("NEW")))
			lines = append(lines, "")
			for _, item := range d.queue.NewItems {
				lines = append(lines, d.renderQueueItem(item, itemIdx, w))
				itemIdx++
			}
		}

		lines = append(lines, "")
		lines = append(lines,
			lipgloss.NewStyle().Foreground(theme.Current.TextMuted).
				Render(fmt.Sprintf("  Total: %d  (enter=open, r=review)", d.queue.Total())))
	}

	return strings.Join(lines, "\n")
}

func (d *Dashboard) renderQueueItem(item study.QueueItem, idx, w int) string {
	diff := styles.DifficultyStyle(item.Difficulty).Render(fmt.Sprintf("[%s]", item.Difficulty[:3]))
	text := "#" + item.FrontendID + " " + truncate(item.Title, w-20) + " " + diff

	if idx == d.cursor && d.focused {
		return queueSelectedStyle.Width(w - 2).Render("▌ " + text)
	}
	return queueItemStyle.Render("  " + text)
}

func (d *Dashboard) renderStats(w int) string {
	var lines []string
	lines = append(lines, dashTitleStyle.Width(w).Render("Stats"))

	if d.stats == nil {
		lines = append(lines, "Loading...")
		return strings.Join(lines, "\n")
	}

	solved := lipgloss.NewStyle().Foreground(theme.Current.Success).Bold(true)
	lines = append(lines, solved.Render(fmt.Sprintf("Solved: %d", d.stats.TotalSolved)))
	lines = append(lines, "")

	barW := 15
	easyTotal := d.stats.EasyTotal
	if easyTotal == 0 {
		easyTotal = int64(d.stats.EasySolved) + 10
	}
	mediumTotal := d.stats.MediumTotal
	if mediumTotal == 0 {
		mediumTotal = int64(d.stats.MediumSolved) + 10
	}
	hardTotal := d.stats.HardTotal
	if hardTotal == 0 {
		hardTotal = int64(d.stats.HardSolved) + 10
	}
	lines = append(lines, styles.DifficultyStyle("Easy").Render(fmt.Sprintf("Easy:   %d/%d", d.stats.EasySolved, easyTotal))+
		" "+stats.ColoredProgressBar(int(d.stats.EasySolved), int(easyTotal), barW, theme.Current.Easy))
	lines = append(lines, styles.DifficultyStyle("Medium").Render(fmt.Sprintf("Medium: %d/%d", d.stats.MediumSolved, mediumTotal))+
		" "+stats.ColoredProgressBar(int(d.stats.MediumSolved), int(mediumTotal), barW, theme.Current.Medium))
	lines = append(lines, styles.DifficultyStyle("Hard").Render(fmt.Sprintf("Hard:   %d/%d", d.stats.HardSolved, hardTotal))+
		" "+stats.ColoredProgressBar(int(d.stats.HardSolved), int(hardTotal), barW, theme.Current.Hard))
	lines = append(lines, "")

	reviewDone := lipgloss.NewStyle().Foreground(theme.Current.Secondary)
	lines = append(lines, reviewDone.Render(fmt.Sprintf("Reviews today: %d", d.stats.ReviewsDone)))

	if d.stats.ActivePlan != "" {
		lines = append(lines, "")
		lines = append(lines, dashSectionStyle.Render("Active Plan:"))
		lines = append(lines, fmt.Sprintf("  %s", d.stats.ActivePlan))
		pct := d.stats.PlanProgress.Percent()
		lines = append(lines, fmt.Sprintf("  %.0f%% complete (%d/%d)",
			pct, d.stats.PlanProgress.Completed, d.stats.PlanProgress.Total))
		lines = append(lines, "  "+stats.ProgressBar(d.stats.PlanProgress.Completed, d.stats.PlanProgress.Total, w-10))
	}

	return strings.Join(lines, "\n")
}

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-2]) + ".."
}
