package srs

import (
	"math"
	"time"
)

// FSRS implements the Free Spaced Repetition Scheduler algorithm (FSRS-5).
// Reference: https://github.com/open-spaced-repetition/fsrs4anki
type FSRS struct {
	Weights       []float64
	RetentionRate float64 // target retention, default 0.9
	MaxInterval   float64 // max interval in days, default 36500
}

// DefaultFSRS creates an FSRS instance with default FSRS-5 parameters.
func DefaultFSRS() *FSRS {
	return &FSRS{
		Weights: []float64{
			0.4072, 1.1829, 3.1262, 15.4722,
			7.2102,
			0.5316, 1.0651, 0.0589,
			1.5330, 0.1544, 1.0347,
			1.9395, 0.1100, 0.2939,
			2.0091, 0.2640, 2.9898,
			0.5100, 0.6100,
		},
		RetentionRate: 0.9,
		MaxInterval:   36500,
	}
}

// NewFSRS creates an FSRS instance with custom weights and retention.
func NewFSRS(weights []float64, retention float64) *FSRS {
	f := DefaultFSRS()
	if len(weights) >= 19 {
		f.Weights = weights
	}
	if retention > 0 && retention < 1 {
		f.RetentionRate = retention
	}
	return f
}

// Schedule computes scheduling info for all 4 ratings given a card and current time.
func (f *FSRS) Schedule(card Card, now time.Time) map[Rating]SchedulingInfo {
	elapsedDays := 0.0
	if !card.LastReview.IsZero() {
		elapsedDays = now.Sub(card.LastReview).Hours() / 24.0
	}

	results := make(map[Rating]SchedulingInfo)

	for _, rating := range []Rating{Again, Hard, Good, Easy} {
		newCard := card
		newCard.ElapsedDays = elapsedDays
		newCard.LastReview = now
		newCard.Reps++

		switch card.State {
		case New:
			f.scheduleNew(&newCard, rating)
		case Learning, Relearning:
			f.scheduleLearning(&newCard, rating, elapsedDays)
		case Review:
			f.scheduleReview(&newCard, rating, elapsedDays)
		}

		// Clamp interval
		if newCard.ScheduledDays > f.MaxInterval {
			newCard.ScheduledDays = f.MaxInterval
		}
		newCard.Due = now.Add(time.Duration(newCard.ScheduledDays*24) * time.Hour)

		results[rating] = SchedulingInfo{
			Card: newCard,
			Log: ReviewLog{
				ProblemID:     card.ProblemID,
				Rating:        rating,
				State:         card.State,
				ElapsedDays:   elapsedDays,
				ScheduledDays: newCard.ScheduledDays,
				ReviewedAt:    now,
			},
		}
	}

	return results
}

func (f *FSRS) scheduleNew(card *Card, rating Rating) {
	card.Difficulty = f.initDifficulty(rating)
	card.Stability = f.initStability(rating)

	switch rating {
	case Again:
		card.State = Learning
		card.ScheduledDays = 0 // review again same session
		card.Lapses++
	case Hard:
		card.State = Learning
		card.ScheduledDays = 0
	case Good:
		card.State = Learning
		card.ScheduledDays = 0
	case Easy:
		card.State = Review
		card.ScheduledDays = f.nextInterval(card.Stability)
	}
}

func (f *FSRS) scheduleLearning(card *Card, rating Rating, elapsedDays float64) {
	switch rating {
	case Again:
		card.State = Learning
		card.Stability = f.initStability(Again)
		card.ScheduledDays = 0
		card.Lapses++
	case Hard:
		card.State = Learning
		card.ScheduledDays = 0
	case Good:
		card.State = Review
		card.Stability = f.initStability(Good)
		card.ScheduledDays = f.nextInterval(card.Stability)
	case Easy:
		card.State = Review
		card.Stability = f.initStability(Easy)
		card.ScheduledDays = f.nextInterval(card.Stability)
	}

	card.Difficulty = f.nextDifficulty(card.Difficulty, rating)
}

func (f *FSRS) scheduleReview(card *Card, rating Rating, elapsedDays float64) {
	retrievability := f.retrievability(elapsedDays, card.Stability)
	card.Difficulty = f.nextDifficulty(card.Difficulty, rating)

	switch rating {
	case Again:
		card.State = Relearning
		card.Stability = f.nextForgetStability(card.Difficulty, card.Stability, retrievability)
		card.ScheduledDays = 0
		card.Lapses++
	case Hard:
		card.State = Review
		card.Stability = f.nextRecallStability(card.Difficulty, card.Stability, retrievability, Hard)
		card.ScheduledDays = f.nextInterval(card.Stability)
	case Good:
		card.State = Review
		card.Stability = f.nextRecallStability(card.Difficulty, card.Stability, retrievability, Good)
		card.ScheduledDays = f.nextInterval(card.Stability)
	case Easy:
		card.State = Review
		card.Stability = f.nextRecallStability(card.Difficulty, card.Stability, retrievability, Easy)
		card.ScheduledDays = f.nextInterval(card.Stability)
	}
}

// initStability returns initial stability for a given rating (w0-w3).
func (f *FSRS) initStability(rating Rating) float64 {
	return f.Weights[int(rating)-1]
}

// initDifficulty returns initial difficulty based on rating.
// D0(G) = w4 - e^(w5*(G-1)) + 1
func (f *FSRS) initDifficulty(rating Rating) float64 {
	d := f.Weights[4] - math.Exp(f.Weights[5]*float64(rating-1)) + 1
	return clamp(d, 1, 10)
}

// nextDifficulty updates difficulty after a review.
// D'(D,G) = w7 * D0(G) + (1-w7) * (D - w6*(G-3))
func (f *FSRS) nextDifficulty(d float64, rating Rating) float64 {
	d0 := f.initDifficulty(rating)
	newD := f.Weights[7]*d0 + (1-f.Weights[7])*(d-f.Weights[6]*float64(rating-3))
	return clamp(newD, 1, 10)
}

// retrievability computes the probability of recall.
// R(t, S) = (1 + t/(9*S))^(-0.5)  [power forgetting curve]
func (f *FSRS) retrievability(elapsedDays, stability float64) float64 {
	if stability <= 0 {
		return 0
	}
	return math.Pow(1+elapsedDays/(9*stability), -0.5)
}

// nextInterval calculates the next review interval from stability.
// I(r, S) = 9 * S * (1/r - 1)
func (f *FSRS) nextInterval(stability float64) float64 {
	r := f.RetentionRate
	interval := 9 * stability * (1/r - 1)
	return math.Max(1, math.Round(interval))
}

// nextRecallStability computes new stability after successful recall.
// S'(D, S, R, G) = S * (e^(w8) * (11-D) * S^(-w9) * (e^(w10*(1-R))-1) * hardPenalty * easyBonus + 1)
func (f *FSRS) nextRecallStability(d, s, r float64, rating Rating) float64 {
	hardPenalty := 1.0
	if rating == Hard {
		hardPenalty = f.Weights[15]
	}
	easyBonus := 1.0
	if rating == Easy {
		easyBonus = f.Weights[16]
	}

	newS := s * (math.Exp(f.Weights[8])*
		(11-d)*
		math.Pow(s, -f.Weights[9])*
		(math.Exp(f.Weights[10]*(1-r))-1)*
		hardPenalty*
		easyBonus + 1)

	return math.Max(0.01, newS)
}

// nextForgetStability computes new stability after forgetting.
// S'(D, S, R) = w11 * D^(-w12) * ((S+1)^w13 - 1) * e^(w14*(1-R))
func (f *FSRS) nextForgetStability(d, s, r float64) float64 {
	newS := f.Weights[11] *
		math.Pow(d, -f.Weights[12]) *
		(math.Pow(s+1, f.Weights[13]) - 1) *
		math.Exp(f.Weights[14]*(1-r))
	return math.Max(0.01, math.Min(newS, s))
}

// Retrievability returns the current recall probability for a card.
func (f *FSRS) Retrievability(card Card, now time.Time) float64 {
	if card.State == New || card.LastReview.IsZero() {
		return 0
	}
	elapsed := now.Sub(card.LastReview).Hours() / 24.0
	return f.retrievability(elapsed, card.Stability)
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
