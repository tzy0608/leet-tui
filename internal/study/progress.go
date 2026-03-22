package study

import (
	"context"
	"database/sql"

	"github.com/leet-tui/leet-tui/internal/db/sqlc"
)

// Stats holds overall progress statistics.
type Stats struct {
	TotalSolved   int64
	EasySolved    int64
	MediumSolved  int64
	HardSolved    int64
	EasyTotal     int64
	MediumTotal   int64
	HardTotal     int64
	ReviewsDone   int64
	ActivePlan    string
	PlanProgress  PlanProgress
	Streak        int
}

// GetStats computes overall statistics.
func GetStats(ctx context.Context, queries *sqlc.Queries, site ...string) (*Stats, error) {
	stats := &Stats{}

	siteVal := "us"
	if len(site) > 0 && site[0] != "" {
		siteVal = site[0]
	}
	siteParam := sql.NullString{String: siteVal, Valid: true}

	total, err := queries.CountSolvedProblems(ctx, siteParam)
	if err != nil {
		return nil, err
	}
	stats.TotalSolved = total

	// Get solved problems by difficulty
	solvedByDiff, err := queries.CountSolvedByDifficulty(ctx, siteParam)
	if err == nil {
		for _, row := range solvedByDiff {
			switch row.Difficulty {
			case "Easy":
				stats.EasySolved = row.Cnt
			case "Medium":
				stats.MediumSolved = row.Cnt
			case "Hard":
				stats.HardSolved = row.Cnt
			}
		}
	}

	// Get total problems by difficulty
	diffCounts, err := queries.CountProblemsByDifficultyAll(ctx, siteParam)
	if err == nil {
		for _, dc := range diffCounts {
			switch dc.Difficulty {
			case "Easy":
				stats.EasyTotal = dc.Cnt
			case "Medium":
				stats.MediumTotal = dc.Cnt
			case "Hard":
				stats.HardTotal = dc.Cnt
			}
		}
	}

	reviewsToday, err := queries.CountReviewsToday(ctx)
	if err != nil {
		return nil, err
	}
	stats.ReviewsDone = reviewsToday

	// Get active plan progress
	activePlan, err := queries.GetActivePlan(ctx)
	if err == nil {
		stats.ActivePlan = activePlan.Name
		progress, err := queries.CountPlanProgress(ctx, sql.NullInt64{Int64: activePlan.ID, Valid: true})
		if err == nil {
			stats.PlanProgress = PlanProgress{
				Total:     int(progress.Total),
				Completed: int(progress.Completed.Float64),
			}
		}
	}

	return stats, nil
}
