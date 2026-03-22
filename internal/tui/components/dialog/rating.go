package dialog

import (
	"fmt"
	"math"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tzy0608/leet-tui/internal/srs"
	"github.com/tzy0608/leet-tui/internal/tui/theme"
)

// RatingMsg is sent when the user selects a rating.
type RatingMsg struct {
	Rating srs.Rating
}

// Rating is the SRS rating selection dialog.
type RatingDialog struct {
	cursor    int // 0-3 → Again/Hard/Good/Easy
	intervals [4]float64
	fsrs      *srs.FSRS
	card      srs.Card
}

var ratingLabels = [4]string{"Again", "Hard", "Good", "Easy"}
var ratingColors = [4]lipgloss.AdaptiveColor{
	{Light: "#D20F39", Dark: "#F38BA8"}, // Again - red
	{Light: "#DF8E1D", Dark: "#F9E2AF"}, // Hard - yellow
	{Light: "#179299", Dark: "#94E2D5"}, // Good - teal
	{Light: "#40A02B", Dark: "#A6E3A1"}, // Easy - green
}

// NewRatingDialog creates a new rating dialog with computed intervals.
func NewRatingDialog(f *srs.FSRS, card srs.Card) *RatingDialog {
	d := &RatingDialog{
		cursor: 2, // default to Good
		fsrs:   f,
		card:   card,
	}

	results := f.Schedule(card, card.Due)
	for i, rating := range []srs.Rating{srs.Again, srs.Hard, srs.Good, srs.Easy} {
		d.intervals[i] = results[rating].Card.ScheduledDays
	}

	return d
}

// Update handles keyboard input.
func (d *RatingDialog) Update(msg tea.Msg) (*RatingDialog, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "left", "h":
			if d.cursor > 0 {
				d.cursor--
			}
		case "right", "l":
			if d.cursor < 3 {
				d.cursor++
			}
		case "1":
			d.cursor = 0
			return d, d.emitRating()
		case "2":
			d.cursor = 1
			return d, d.emitRating()
		case "3":
			d.cursor = 2
			return d, d.emitRating()
		case "4":
			d.cursor = 3
			return d, d.emitRating()
		case "enter":
			return d, d.emitRating()
		}
	}
	return d, nil
}

func (d *RatingDialog) emitRating() tea.Cmd {
	rating := srs.Rating(d.cursor + 1)
	return func() tea.Msg { return RatingMsg{Rating: rating} }
}

// View renders the rating dialog.
func (d *RatingDialog) View() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Current.Primary).
		Render("How well did you remember?")

	hint := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#6C7086", Dark: "#6C7086"}).
		Render("← → to select, Enter to confirm, 1-4 to quick rate")

	var buttons []string
	for i, label := range ratingLabels {
		interval := d.intervals[i]
		nextStr := formatInterval(interval)

		col := lipgloss.NewStyle().
			Foreground(ratingColors[i])

		var btnStyle lipgloss.Style
		if d.cursor == i {
			btnStyle = lipgloss.NewStyle().
				Background(ratingColors[i]).
				Foreground(lipgloss.Color("#1E1E2E")).
				Bold(true).
				Width(12).
				Align(lipgloss.Center).
				Padding(0, 1).
				MarginRight(1)
		} else {
			btnStyle = lipgloss.NewStyle().
				Background(lipgloss.AdaptiveColor{Light: "#E6E9EF", Dark: "#313244"}).
				Foreground(ratingColors[i]).
				Width(12).
				Align(lipgloss.Center).
				Padding(0, 1).
				MarginRight(1)
		}

		labelLine := fmt.Sprintf("[%d] %s", i+1, label)
		nextLine := col.Render(nextStr)
		btn := btnStyle.Render(labelLine + "\n" + nextLine)
		buttons = append(buttons, btn)
	}

	buttonRow := lipgloss.JoinHorizontal(lipgloss.Top, buttons...)

	content := title + "\n\n" + buttonRow + "\n\n" + hint

	return dialogStyle.Width(56).Render(content)
}

func formatInterval(days float64) string {
	if days == 0 {
		return "now"
	}
	if days < 1 {
		return fmt.Sprintf("%dm", int(days*60*24))
	}
	if days < 30 {
		return fmt.Sprintf("%dd", int(math.Round(days)))
	}
	if days < 365 {
		return fmt.Sprintf("%.1fm", days/30)
	}
	return fmt.Sprintf("%.1fy", days/365)
}
