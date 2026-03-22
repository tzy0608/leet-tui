package srs

import (
	"database/sql"
	"time"

	"github.com/leet-tui/leet-tui/internal/db/sqlc"
)

// Scheduler manages the review queue and processes reviews.
type Scheduler struct {
	fsrs    *FSRS
	queries *sqlc.Queries
}

// NewScheduler creates a new SRS scheduler.
func NewScheduler(fsrs *FSRS, queries *sqlc.Queries) *Scheduler {
	return &Scheduler{fsrs: fsrs, queries: queries}
}

// DueItem represents a problem due for review with metadata.
type DueItem struct {
	ProblemID  int
	Title      string
	FrontendID string
	Difficulty string
	TitleSlug  string
	Card       Card
}

// GetDueItems returns all review cards that are due.
func (s *Scheduler) GetDueItems(ctx any, now time.Time) ([]DueItem, error) {
	// This would query the database for due cards
	// Implementation depends on sqlc generated code
	return nil, nil
}

// ProcessReview updates a card based on the user's rating.
func (s *Scheduler) ProcessReview(card Card, rating Rating, timeSpentSec int, now time.Time) (SchedulingInfo, error) {
	results := s.fsrs.Schedule(card, now)
	info := results[rating]
	info.Log.TimeSpentSec = timeSpentSec
	return info, nil
}

// CreateCard creates a new review card for a problem.
func NewCard(problemID int, now time.Time) Card {
	return Card{
		ProblemID: problemID,
		Due:       now,
		State:     New,
	}
}

// CardFromDB converts a database row to a Card.
func CardFromDB(row sqlc.ReviewCard) Card {
	card := Card{
		ProblemID:     int(row.ProblemID),
		Stability:     row.Stability,
		Difficulty:    row.Difficulty,
		ElapsedDays:   row.ElapsedDays,
		ScheduledDays: row.ScheduledDays,
		Reps:          int(row.Reps),
		Lapses:        int(row.Lapses),
		State:         State(row.State),
	}

	card.Due = row.Due

	if row.LastReview.Valid {
		card.LastReview = row.LastReview.Time
	}

	return card
}

// CardToDBParams converts a Card to database upsert parameters.
func CardToDBParams(card Card) sqlc.UpsertReviewCardParams {
	params := sqlc.UpsertReviewCardParams{
		ProblemID:     int64(card.ProblemID),
		Due:           card.Due,
		Stability:     card.Stability,
		Difficulty:    card.Difficulty,
		ElapsedDays:   card.ElapsedDays,
		ScheduledDays: card.ScheduledDays,
		Reps:          int64(card.Reps),
		Lapses:        int64(card.Lapses),
		State:         int64(card.State),
	}

	if !card.LastReview.IsZero() {
		params.LastReview = sql.NullTime{
			Time:  card.LastReview,
			Valid: true,
		}
	}

	return params
}

// SuggestRating suggests a rating based on time spent and whether hints were used.
func SuggestRating(timeSpentSec int, usedHint bool, difficulty string) Rating {
	// Quick solve without hints → Easy
	// Normal solve → Good
	// Slow/with hints → Hard
	// Couldn't solve → Again

	thresholdSec := 300 // 5 minutes default
	switch difficulty {
	case "Easy":
		thresholdSec = 180
	case "Medium":
		thresholdSec = 420
	case "Hard":
		thresholdSec = 600
	}

	if usedHint {
		if timeSpentSec > thresholdSec {
			return Again
		}
		return Hard
	}

	if timeSpentSec < thresholdSec/3 {
		return Easy
	}
	if timeSpentSec < thresholdSec {
		return Good
	}
	return Hard
}
