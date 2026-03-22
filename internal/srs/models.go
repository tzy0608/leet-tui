package srs

import "time"

// State represents the learning state of a card.
type State int

const (
	New        State = 0
	Learning   State = 1
	Review     State = 2
	Relearning State = 3
)

func (s State) String() string {
	switch s {
	case New:
		return "New"
	case Learning:
		return "Learning"
	case Review:
		return "Review"
	case Relearning:
		return "Relearning"
	default:
		return "Unknown"
	}
}

// Rating represents how well the user recalled a card.
type Rating int

const (
	Again Rating = 1
	Hard  Rating = 2
	Good  Rating = 3
	Easy  Rating = 4
)

func (r Rating) String() string {
	switch r {
	case Again:
		return "Again"
	case Hard:
		return "Hard"
	case Good:
		return "Good"
	case Easy:
		return "Easy"
	default:
		return "Unknown"
	}
}

// Card holds the SRS state for a single problem.
type Card struct {
	ProblemID     int
	Due           time.Time
	Stability     float64
	Difficulty    float64
	ElapsedDays   float64
	ScheduledDays float64
	Reps          int
	Lapses        int
	State         State
	LastReview    time.Time
}

// ReviewLog records a single review event.
type ReviewLog struct {
	ProblemID    int
	Rating       Rating
	State        State
	ElapsedDays  float64
	ScheduledDays float64
	TimeSpentSec int
	ReviewedAt   time.Time
}

// SchedulingInfo holds the result of scheduling for each rating.
type SchedulingInfo struct {
	Card Card
	Log  ReviewLog
}
