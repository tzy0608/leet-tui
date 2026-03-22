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

// PlansLoadedMsg provides plan data.
type PlansLoadedMsg struct {
	Plans    []study.Plan
	Progress map[int]study.PlanProgress
}

// PlanSelectedMsg is sent when a plan is activated.
type PlanSelectedMsg struct {
	PlanID int
}

// Plans is the study plan management page.
type Plans struct {
	plans    []study.Plan
	progress map[int]study.PlanProgress
	cursor   int
	focused  bool
	width    int
	height   int
}

// NewPlans creates a new plans page.
func NewPlans() *Plans {
	return &Plans{
		progress: make(map[int]study.PlanProgress),
	}
}

// SetData updates the plans page with loaded data.
func (p *Plans) SetData(msg PlansLoadedMsg) {
	p.plans = msg.Plans
	p.progress = msg.Progress
	if p.cursor >= len(p.plans) {
		p.cursor = 0
	}
}

func (p *Plans) Init() tea.Cmd { return nil }

func (p *Plans) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case PlansLoadedMsg:
		p.SetData(msg)

	case tea.KeyMsg:
		if !p.focused {
			return p, nil
		}
		switch msg.String() {
		case "j", "down":
			if p.cursor < len(p.plans)-1 {
				p.cursor++
			}
		case "k", "up":
			if p.cursor > 0 {
				p.cursor--
			}
		case "enter", "a":
			if p.cursor < len(p.plans) {
				id := p.plans[p.cursor].ID
				return p, func() tea.Msg { return PlanSelectedMsg{PlanID: id} }
			}
		}
	}
	return p, nil
}

func (p *Plans) Focus() tea.Cmd { p.focused = true; return nil }
func (p *Plans) Blur() tea.Cmd  { p.focused = false; return nil }
func (p *Plans) IsFocused() bool { return p.focused }

func (p *Plans) SetSize(w, h int) tea.Cmd {
	p.width = w
	p.height = h
	return nil
}

func (p *Plans) GetSize() (int, int) { return p.width, p.height }

func (p *Plans) BindingKeys() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/↓", "next plan")),
		key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k/↑", "prev plan")),
		key.NewBinding(key.WithKeys("enter", "a"), key.WithHelp("enter", "activate plan")),
	}
}

func (p *Plans) View() string {
	if p.width == 0 {
		return "Loading..."
	}

	panelH := p.height - 3 // hint line + outer border top/bottom
	listW := (p.width - 2) / 3
	detailW := p.width - 2 - listW - 1

	listView := p.renderList(listW)
	detailView := p.renderDetail(detailW)

	leftCol := lipgloss.NewStyle().Width(listW).Height(panelH).Padding(1, 2).Render(listView)
	rightCol := lipgloss.NewStyle().Width(detailW).Height(panelH).Padding(1, 2).Render(detailView)
	divider := styles.VerticalDivider(panelH)

	inner := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, divider, rightCol)
	outer := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current.Border).
		Render(inner)

	hint := lipgloss.NewStyle().
		Foreground(theme.Current.TextMuted).
		Render("  j/k Navigate  Enter Activate plan")

	return outer + "\n" + hint
}

func (p *Plans) renderList(w int) string {
	title := styles.PanelHeader.Width(w - 6).Render("Study Plans")
	var rows []string
	rows = append(rows, title)

	planSelectedStyle := lipgloss.NewStyle().
		Background(lipgloss.AdaptiveColor{Light: "#DDD6FE", Dark: "#3B2D63"}).
		Foreground(theme.Current.Primary).
		Width(w - 6).
		Padding(0, 1)

	planNormalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#4C4F69", Dark: "#CDD6F4"}).
		Width(w - 6).
		Padding(0, 1)

	activeBadge := styles.NewBadge(
		lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#1E1E2E"},
		theme.Current.Success,
	)

	if len(p.plans) == 0 {
		rows = append(rows, "")
		rows = append(rows, lipgloss.NewStyle().Foreground(theme.Current.TextMuted).Render("  No plans yet"))
		return strings.Join(rows, "\n")
	}

	for i, plan := range p.plans {
		name := plan.Name
		maxNameW := w - 10 // reserve space for prefix + badge
		nameRunes := []rune(name)
		if lipgloss.Width(name) > maxNameW {
			// Truncate by runes, then check width
			for len(nameRunes) > 0 && lipgloss.Width(string(nameRunes)) > maxNameW-2 {
				nameRunes = nameRunes[:len(nameRunes)-1]
			}
			name = string(nameRunes) + ".."
		}

		var label string
		if plan.IsActive {
			label = activeBadge.Render("Active") + " " + name
		} else {
			label = "  " + name
		}

		if i == p.cursor && p.focused {
			rows = append(rows, planSelectedStyle.Render("▌ "+label))
		} else {
			rows = append(rows, planNormalStyle.Render("  "+label))
		}
	}

	return strings.Join(rows, "\n")
}

func (p *Plans) renderDetail(w int) string {
	if len(p.plans) == 0 || p.cursor >= len(p.plans) {
		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Current.Primary)
		muted := lipgloss.NewStyle().Foreground(theme.Current.TextMuted)
		secondary := lipgloss.NewStyle().Foreground(theme.Current.Secondary)
		var guide []string
		guide = append(guide, titleStyle.Render("Getting Started"))
		guide = append(guide, "")
		guide = append(guide, muted.Render("This app includes curated study plans:"))
		guide = append(guide, secondary.Render("  • Blind 75 — Classic interview prep"))
		guide = append(guide, secondary.Render("  • NeetCode 150 — Comprehensive patterns"))
		guide = append(guide, "")
		guide = append(guide, muted.Render("To get started:"))
		guide = append(guide, muted.Render("  1. Go to Problems (press 2) and sync"))
		guide = append(guide, muted.Render("  2. Return here — plans appear automatically"))
		guide = append(guide, muted.Render("  3. Select a plan and press Enter to activate"))
		guide = append(guide, "")
		guide = append(guide, muted.Render("Tip: Active plan problems show up in your"))
		guide = append(guide, muted.Render("daily queue on the Dashboard."))
		return strings.Join(guide, "\n")
	}

	plan := p.plans[p.cursor]
	var lines []string

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Current.Primary)
	lines = append(lines, titleStyle.Render(plan.Name))
	lines = append(lines, "")

	if plan.Description != "" {
		lines = append(lines, lipgloss.NewStyle().Foreground(theme.Current.TextMuted).Render(plan.Description))
		lines = append(lines, "")
	}

	if plan.IsPredefined {
		lines = append(lines, lipgloss.NewStyle().Foreground(theme.Current.Secondary).Render("✓ Curated plan"))
	}

	if plan.IsActive {
		badge := styles.NewBadge(
			lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#1E1E2E"},
			theme.Current.Success,
		)
		lines = append(lines, badge.Render("Active"))
	}

	// Progress
	if prog, ok := p.progress[plan.ID]; ok && prog.Total > 0 {
		lines = append(lines, "")
		pct := prog.Percent()
		lines = append(lines, fmt.Sprintf("Progress: %d/%d (%.0f%%)", prog.Completed, prog.Total, pct))

		barW := w - 12
		if barW < 10 {
			barW = 10
		}
		lines = append(lines, stats.ProgressBar(prog.Completed, prog.Total, barW))
	}

	lines = append(lines, "")
	if plan.IsActive {
		lines = append(lines, lipgloss.NewStyle().Foreground(theme.Current.TextMuted).
			Render("This plan is active. New problems will appear in your daily queue."))
	} else {
		lines = append(lines, lipgloss.NewStyle().Foreground(theme.Current.TextMuted).Render("Enter to activate this plan"))
	}

	return strings.Join(lines, "\n")
}
