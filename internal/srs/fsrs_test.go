package srs

import (
	"math"
	"testing"
	"time"
)

func TestInitStability(t *testing.T) {
	f := DefaultFSRS()

	tests := []struct {
		rating   Rating
		expected float64
	}{
		{Again, 0.4072},
		{Hard, 1.1829},
		{Good, 3.1262},
		{Easy, 15.4722},
	}

	for _, tt := range tests {
		got := f.initStability(tt.rating)
		if math.Abs(got-tt.expected) > 0.001 {
			t.Errorf("initStability(%v) = %v, want %v", tt.rating, got, tt.expected)
		}
	}
}

func TestInitDifficulty(t *testing.T) {
	f := DefaultFSRS()

	// D0(G) = w4 - e^(w5*(G-1)) + 1
	// For Good (3): w4 - e^(w5*2) + 1
	d := f.initDifficulty(Good)
	if d < 1 || d > 10 {
		t.Errorf("initDifficulty(Good) = %v, want [1,10]", d)
	}

	// Easy should have lower difficulty than Again
	dEasy := f.initDifficulty(Easy)
	dAgain := f.initDifficulty(Again)
	if dEasy >= dAgain {
		t.Errorf("Easy difficulty (%v) should be less than Again (%v)", dEasy, dAgain)
	}
}

func TestRetrievability(t *testing.T) {
	f := DefaultFSRS()

	// At t=0, R should be 1.0
	r0 := f.retrievability(0, 10)
	if math.Abs(r0-1.0) > 0.001 {
		t.Errorf("R(0, 10) = %v, want 1.0", r0)
	}

	// R should decrease over time
	r1 := f.retrievability(1, 10)
	r10 := f.retrievability(10, 10)
	if r1 <= r10 {
		t.Errorf("R(1) = %v should be > R(10) = %v", r1, r10)
	}

	// Higher stability → slower forgetting
	rLowS := f.retrievability(10, 5)
	rHighS := f.retrievability(10, 50)
	if rHighS <= rLowS {
		t.Errorf("R with S=50 (%v) should be > R with S=5 (%v)", rHighS, rLowS)
	}
}

func TestNextInterval(t *testing.T) {
	f := DefaultFSRS()

	// I = 9 * S * (1/r - 1), with r=0.9
	// For S=1: I = 9 * 1 * (1/0.9 - 1) = 9 * 0.111 = 1
	i := f.nextInterval(1)
	if i != 1 {
		t.Errorf("nextInterval(1) = %v, want 1", i)
	}

	// Higher stability → longer interval
	i10 := f.nextInterval(10)
	i100 := f.nextInterval(100)
	if i100 <= i10 {
		t.Errorf("interval(S=100) = %v should be > interval(S=10) = %v", i100, i10)
	}
}

func TestScheduleNewCard(t *testing.T) {
	f := DefaultFSRS()
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	card := NewCard(1, now)

	results := f.Schedule(card, now)

	// Again should stay in Learning
	if results[Again].Card.State != Learning {
		t.Errorf("Again: state = %v, want Learning", results[Again].Card.State)
	}

	// Easy should go to Review
	if results[Easy].Card.State != Review {
		t.Errorf("Easy: state = %v, want Review", results[Easy].Card.State)
	}

	// Easy should have a positive interval
	if results[Easy].Card.ScheduledDays <= 0 {
		t.Errorf("Easy: scheduledDays = %v, want > 0", results[Easy].Card.ScheduledDays)
	}
}

func TestScheduleReviewCard(t *testing.T) {
	f := DefaultFSRS()
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	card := Card{
		ProblemID:     1,
		Stability:     10,
		Difficulty:    5,
		State:         Review,
		Reps:          3,
		LastReview:    now.Add(-24 * time.Hour), // reviewed 1 day ago
		ScheduledDays: 10,
	}

	results := f.Schedule(card, now)

	// Again should go to Relearning
	if results[Again].Card.State != Relearning {
		t.Errorf("Again: state = %v, want Relearning", results[Again].Card.State)
	}

	// Again should increase lapses
	if results[Again].Card.Lapses != card.Lapses+1 {
		t.Errorf("Again: lapses = %v, want %v", results[Again].Card.Lapses, card.Lapses+1)
	}

	// Good should stay in Review with increased stability
	goodResult := results[Good]
	if goodResult.Card.State != Review {
		t.Errorf("Good: state = %v, want Review", goodResult.Card.State)
	}

	// Easy should have longer interval than Good
	if results[Easy].Card.ScheduledDays <= results[Good].Card.ScheduledDays {
		t.Errorf("Easy interval (%v) should be > Good interval (%v)",
			results[Easy].Card.ScheduledDays, results[Good].Card.ScheduledDays)
	}
}

func TestNextRecallStability(t *testing.T) {
	f := DefaultFSRS()

	s := 10.0
	d := 5.0
	r := 0.9

	// Good recall should increase stability
	sGood := f.nextRecallStability(d, s, r, Good)
	if sGood <= s {
		t.Errorf("Good recall: new S (%v) should be > old S (%v)", sGood, s)
	}

	// Easy should increase more than Good
	sEasy := f.nextRecallStability(d, s, r, Easy)
	if sEasy <= sGood {
		t.Errorf("Easy recall: S (%v) should be > Good S (%v)", sEasy, sGood)
	}
}

func TestNextForgetStability(t *testing.T) {
	f := DefaultFSRS()

	s := 10.0
	d := 5.0
	r := 0.5

	newS := f.nextForgetStability(d, s, r)
	// After forgetting, stability should decrease
	if newS >= s {
		t.Errorf("Forget: new S (%v) should be < old S (%v)", newS, s)
	}
	if newS <= 0 {
		t.Errorf("Forget: new S (%v) should be > 0", newS)
	}
}

func TestSuggestRating(t *testing.T) {
	tests := []struct {
		timeSec    int
		hint       bool
		difficulty string
		expected   Rating
	}{
		{30, false, "Easy", Easy},     // quick solve, easy problem
		{120, false, "Easy", Good},    // normal solve
		{300, false, "Easy", Hard},    // slow solve
		{60, true, "Easy", Hard},      // with hint, quick
		{300, true, "Easy", Again},    // with hint, slow
	}

	for _, tt := range tests {
		got := SuggestRating(tt.timeSec, tt.hint, tt.difficulty)
		if got != tt.expected {
			t.Errorf("SuggestRating(%d, %v, %q) = %v, want %v",
				tt.timeSec, tt.hint, tt.difficulty, got, tt.expected)
		}
	}
}

func TestClamp(t *testing.T) {
	if clamp(5, 1, 10) != 5 {
		t.Error("clamp(5, 1, 10) should be 5")
	}
	if clamp(-1, 1, 10) != 1 {
		t.Error("clamp(-1, 1, 10) should be 1")
	}
	if clamp(15, 1, 10) != 10 {
		t.Error("clamp(15, 1, 10) should be 10")
	}
}
