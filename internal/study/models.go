package study

import "time"

// Plan represents a study plan.
type Plan struct {
	ID           int
	Name         string
	Slug         string
	Description  string
	IsPredefined bool
	IsActive     bool
	CreatedAt    time.Time
}

// PlanProblem represents a problem within a plan.
type PlanProblem struct {
	PlanID      int
	ProblemID   int
	DayNumber   int
	TopicGroup  string
	SortOrder   int
	IsCompleted bool
	CompletedAt time.Time

	// Joined fields
	Title      string
	FrontendID string
	Difficulty string
	TitleSlug  string
}

// PlanProgress tracks plan completion stats.
type PlanProgress struct {
	Total     int
	Completed int
}

func (p PlanProgress) Percent() float64 {
	if p.Total == 0 {
		return 0
	}
	return float64(p.Completed) / float64(p.Total) * 100
}

// DailyQueue holds today's problems (reviews + new).
type DailyQueue struct {
	ReviewItems []QueueItem
	NewItems    []QueueItem
}

// QueueItem is a problem in the daily queue.
type QueueItem struct {
	ProblemID  int
	Title      string
	FrontendID string
	Difficulty string
	TitleSlug  string
	IsReview   bool
	DueAt      time.Time
}

func (q DailyQueue) Total() int {
	return len(q.ReviewItems) + len(q.NewItems)
}

// PlanDefinition is used for loading predefined plans from JSON.
type PlanDefinition struct {
	Name        string              `json:"name"`
	Slug        string              `json:"slug"`
	Description string              `json:"description"`
	Groups      []PlanGroupDef      `json:"groups"`
}

type PlanGroupDef struct {
	Topic    string   `json:"topic"`
	Day      int      `json:"day"`
	Problems []string `json:"problems"` // title_slugs
}
