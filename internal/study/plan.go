package study

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/tzy0608/leet-tui/internal/db/sqlc"
	"github.com/tzy0608/leet-tui/internal/srs"
	"github.com/tzy0608/leet-tui/internal/study/plans"
)

// Engine manages study plans and daily queue generation.
type Engine struct {
	queries   *sqlc.Queries
	scheduler *srs.Scheduler
	newPerDay int
}

// NewEngine creates a new study plan engine.
func NewEngine(queries *sqlc.Queries, scheduler *srs.Scheduler, newPerDay int) *Engine {
	return &Engine{
		queries:   queries,
		scheduler: scheduler,
		newPerDay: newPerDay,
	}
}

// LoadPredefinedPlans loads embedded plan definitions into the database.
func (e *Engine) LoadPredefinedPlans(ctx context.Context) error {
	planFiles := map[string]string{
		"blind75.json":       "blind75",
		"neetcode150.json":   "neetcode150",
		"leetcode21day.json": "leetcode-21day",
	}

	for filename, slug := range planFiles {
		// Check if already exists
		_, err := e.queries.GetStudyPlanBySlug(ctx, slug)
		if err == nil {
			continue // already loaded
		}

		data, err := plans.FS.ReadFile(filename)
		if err != nil {
			continue // file not found, skip
		}

		var def PlanDefinition
		if err := json.Unmarshal(data, &def); err != nil {
			return fmt.Errorf("parse plan %s: %w", filename, err)
		}

		plan, err := e.queries.CreateStudyPlan(ctx, sqlc.CreateStudyPlanParams{
			Name:         def.Name,
			Slug:         def.Slug,
			Description:  sql.NullString{String: def.Description, Valid: true},
			IsPredefined: sql.NullInt64{Int64: 1, Valid: true},
			IsActive:     sql.NullInt64{Int64: 0, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("create plan %s: %w", def.Slug, err)
		}

		for _, group := range def.Groups {
			for i, slug := range group.Problems {
				problem, err := e.queries.GetProblemBySlug(ctx, slug)
				if err != nil {
					continue // problem not synced yet
				}
				e.queries.AddProblemToPlan(ctx, sqlc.AddProblemToPlanParams{
					PlanID:     sql.NullInt64{Int64: plan.ID, Valid: true},
					ProblemID:  sql.NullInt64{Int64: problem.ID, Valid: true},
					DayNumber:  int64(group.Day),
					TopicGroup: sql.NullString{String: group.Topic, Valid: true},
					SortOrder:  sql.NullInt64{Int64: int64(i), Valid: true},
				})
			}
		}
	}

	return nil
}

// RecentAcceptedChecker checks if a problem was recently accepted on LeetCode.
type RecentAcceptedChecker interface {
	FetchRecentAccepted(ctx context.Context, titleSlug string, withinDays int) (bool, error)
}

// SyncPlanCompletion checks LeetCode submission status for incomplete plan problems
// and marks recently accepted ones as completed.
func (e *Engine) SyncPlanCompletion(ctx context.Context, checker RecentAcceptedChecker) error {
	activePlan, err := e.queries.GetActivePlan(ctx)
	if err != nil {
		return nil // no active plan is not an error
	}

	incomplete, err := e.queries.ListIncompletePlanProblems(ctx, sqlc.ListIncompletePlanProblemsParams{
		PlanID: sql.NullInt64{Int64: activePlan.ID, Valid: true},
		Limit:  20,
	})
	if err != nil {
		return fmt.Errorf("list incomplete plan problems: %w", err)
	}

	type result struct {
		problemID int64
		accepted  bool
	}

	syncCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	results := make(chan result, len(incomplete))
	for _, prob := range incomplete {
		wg.Add(1)
		go func(slug string, pid int64) {
			defer wg.Done()
			accepted, err := checker.FetchRecentAccepted(syncCtx, slug, 5)
			if err != nil {
				return
			}
			if accepted {
				results <- result{problemID: pid, accepted: true}
			}
		}(prob.TitleSlug, prob.ProblemID.Int64)
	}
	go func() {
		wg.Wait()
		close(results)
	}()

	for r := range results {
		_ = e.queries.CompletePlanProblem(ctx, sqlc.CompletePlanProblemParams{
			PlanID:    sql.NullInt64{Int64: activePlan.ID, Valid: true},
			ProblemID: sql.NullInt64{Int64: r.problemID, Valid: true},
		})
	}
	return nil
}

// BuildDailyQueue creates today's queue: reviews first, then new problems from active plan.
func (e *Engine) BuildDailyQueue(ctx context.Context, now time.Time) (*DailyQueue, error) {
	queue := &DailyQueue{}

	// 1. Get due review cards
	dueCards, err := e.queries.ListDueReviewCards(ctx, now)
	if err != nil {
		return nil, fmt.Errorf("list due cards: %w", err)
	}

	for _, dc := range dueCards {
		queue.ReviewItems = append(queue.ReviewItems, QueueItem{
			ProblemID:  int(dc.ProblemID),
			Title:      dc.Title,
			FrontendID: dc.FrontendID,
			Difficulty: dc.ProbDifficulty,
			TitleSlug:  dc.TitleSlug,
			IsReview:   true,
		})
	}

	// 2. Get new problems from active plan
	activePlan, err := e.queries.GetActivePlan(ctx)
	if err != nil {
		// No active plan is not an error - just return reviews only
		if err == sql.ErrNoRows {
			return queue, nil
		}
		return queue, fmt.Errorf("get active plan: %w", err)
	}

	incomplete, err := e.queries.ListIncompletePlanProblems(ctx, sqlc.ListIncompletePlanProblemsParams{
		PlanID: sql.NullInt64{Int64: activePlan.ID, Valid: true},
		Limit:  int64(e.newPerDay),
	})
	if err != nil {
		return queue, fmt.Errorf("list incomplete plan problems: %w", err)
	}

	for _, pp := range incomplete {
		queue.NewItems = append(queue.NewItems, QueueItem{
			ProblemID:  int(pp.ProblemID.Int64),
			Title:      pp.Title,
			FrontendID: pp.FrontendID,
			Difficulty: pp.Difficulty,
			TitleSlug:  pp.TitleSlug,
			IsReview:   false,
		})
	}

	return queue, nil
}
