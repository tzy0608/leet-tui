package page

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tzy0608/leet-tui/internal/leetcode"
	"github.com/tzy0608/leet-tui/internal/srs"
	"github.com/tzy0608/leet-tui/internal/tui/components/dialog"
	"github.com/tzy0608/leet-tui/internal/tui/styles"
	"github.com/tzy0608/leet-tui/internal/tui/theme"
)

// ReviewQueueMsg provides the list of due review items.
type ReviewQueueMsg struct {
	Items []ReviewItem
}

// ReviewItem holds a card and its problem info.
type ReviewItem struct {
	Card    srs.Card
	Problem *leetcode.Problem
}

// ReviewDoneMsg is sent when the user rates a problem.
type ReviewDoneMsg struct {
	ProblemID int
	Rating    srs.Rating
	TimeSpent time.Duration
}

// Review is the SRS review session page.
type Review struct {
	items      []ReviewItem
	current    int
	card       *srs.Card
	problem    *leetcode.Problem
	fsrs       *srs.FSRS
	ratingDlg  *dialog.RatingDialog
	showRating bool
	viewport   viewport.Model
	focused    bool
	width      int
	height     int
	startedAt  time.Time
}

// NewReview creates a new review page.
func NewReview(fsrs *srs.FSRS) *Review {
	return &Review{
		fsrs:     fsrs,
		viewport: viewport.New(0, 0),
	}
}

// SetQueue updates the review queue.
func (r *Review) SetQueue(msg ReviewQueueMsg) {
	r.items = msg.Items
	r.current = 0
	r.showRating = false
	r.loadCurrentItem()
}

func (r *Review) loadCurrentItem() {
	if r.current >= len(r.items) {
		r.card = nil
		r.problem = nil
		return
	}
	item := r.items[r.current]
	r.card = &item.Card
	r.problem = item.Problem
	r.startedAt = time.Now()
	r.showRating = false
	r.ratingDlg = nil

	if r.problem != nil {
		descMD := htmlToMarkdown(r.problem.Content)
		// Append last accepted code if available
		if r.problem.LastAcceptedCode != "" {
			descMD += "\n---\n\n**Last Accepted Code:**\n\n```\n" + r.problem.LastAcceptedCode + "\n```\n"
		}
		contentW := r.viewport.Width - 2
		if contentW < 20 {
			contentW = 80
		}
		r.viewport.SetContent(renderMarkdown(descMD, contentW))
	}
}

func (r *Review) Init() tea.Cmd { return nil }

func (r *Review) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case ReviewQueueMsg:
		r.SetQueue(msg)

	case dialog.RatingMsg:
		r.showRating = false
		if r.card != nil {
			spent := time.Since(r.startedAt)
			id := r.card.ProblemID
			rating := msg.Rating

			cmd := func() tea.Msg {
				return ReviewDoneMsg{
					ProblemID: id,
					Rating:    rating,
					TimeSpent: spent,
				}
			}
			cmds = append(cmds, cmd)
		}
		r.current++
		r.loadCurrentItem()

	case tea.KeyMsg:
		if !r.focused {
			return r, nil
		}

		if r.showRating && r.ratingDlg != nil {
			var cmd tea.Cmd
			r.ratingDlg, cmd = r.ratingDlg.Update(msg)
			cmds = append(cmds, cmd)
			return r, tea.Batch(cmds...)
		}

		switch msg.String() {
		case " ", "enter":
			// Show rating dialog
			if r.card != nil {
				r.showRating = true
				r.ratingDlg = dialog.NewRatingDialog(r.fsrs, *r.card)
			}
		case "n":
			// Skip to next without rating
			r.current++
			r.loadCurrentItem()
		case "esc":
			return r, NavigateTo(DashboardPage)
		}
	}

	var cmd tea.Cmd
	r.viewport, cmd = r.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return r, tea.Batch(cmds...)
}

func (r *Review) Focus() tea.Cmd { r.focused = true; return nil }
func (r *Review) Blur() tea.Cmd  { r.focused = false; return nil }
func (r *Review) IsFocused() bool { return r.focused }

func (r *Review) SetSize(w, h int) tea.Cmd {
	r.width = w
	r.height = h
	r.viewport.Width = w - 4
	r.viewport.Height = h - 8
	return nil
}

func (r *Review) GetSize() (int, int) { return r.width, r.height }

func (r *Review) BindingKeys() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("space", "enter"), key.WithHelp("space", "rate problem")),
		key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "skip")),
		key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back to dashboard")),
	}
}

func (r *Review) View() string {
	if r.width == 0 {
		return "Loading..."
	}

	// Queue finished
	if len(r.items) == 0 || r.current >= len(r.items) {
		doneMsg := lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Current.Success).
			Render("All reviews completed for today!")
		hint := lipgloss.NewStyle().
			Foreground(theme.Current.TextMuted).
			Render("\n\nesc to return to dashboard")
		content := styles.Card.Render(doneMsg + hint)
		return lipgloss.Place(r.width, r.height, lipgloss.Center, lipgloss.Center, content)
	}

	// Progress badge
	progressBadge := styles.NewBadge(
		lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#1E1E2E"},
		theme.Current.Secondary,
	)
	progress := progressBadge.Render(fmt.Sprintf("%d/%d", r.current+1, len(r.items)))

	diff := styles.DifficultyStyle(r.problem.Difficulty).Bold(true).Render(r.problem.Difficulty)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Current.Primary)
	title := titleStyle.Render(fmt.Sprintf("#%s %s", r.problem.FrontendID, r.problem.Title))
	header := title + " " + diff + "  " + progress

	// Problem content
	body := styles.Panel.Width(r.width - 2).Height(r.height - 8).Render(r.viewport.View())

	// Instructions or rating dialog
	var bottom string
	if r.showRating && r.ratingDlg != nil {
		bottom = r.ratingDlg.View()
	} else {
		hint := lipgloss.NewStyle().
			Foreground(theme.Current.TextMuted).
			Render("  Space/Enter to rate  n to skip  esc to exit")
		bottom = hint
	}

	content := strings.Join([]string{header, body, bottom}, "\n")

	if r.showRating && r.ratingDlg != nil {
		dlgView := r.ratingDlg.View()
		return lipgloss.Place(r.width, r.height, lipgloss.Center, lipgloss.Center, dlgView)
	}

	return content
}

