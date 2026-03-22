package app

import (
	"context"
	"database/sql"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/leet-tui/leet-tui/internal/config"
	"github.com/leet-tui/leet-tui/internal/db"
	"github.com/leet-tui/leet-tui/internal/db/sqlc"
	"github.com/leet-tui/leet-tui/internal/leetcode"
	"github.com/leet-tui/leet-tui/internal/tui"
)

// App is the main application lifecycle manager.
type App struct {
	cfg *config.Config
}

// New creates a new App.
func New(cfg *config.Config) *App {
	return &App{cfg: cfg}
}

// Run starts the TUI application.
func (a *App) Run(ctx context.Context) error {
	// Open database
	database, err := db.Open(a.cfg.General.DataDir)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer database.Close()

	queries := sqlc.New(database)

	// Create LeetCode client
	lcClient := leetcode.NewClient(
		a.cfg.LeetCode.Site,
		a.cfg.LeetCode.Cookie,
		a.cfg.LeetCode.CSRFToken,
	)

	// Create and run TUI
	model := tui.New(a.cfg, database, queries, lcClient)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
		tea.WithContext(ctx),
	)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("run TUI: %w", err)
	}

	return nil
}

// Sync synchronizes problems from LeetCode to the local database.
func (a *App) Sync(ctx context.Context) error {
	database, err := db.Open(a.cfg.General.DataDir)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer database.Close()

	queries := sqlc.New(database)
	lcClient := leetcode.NewClient(
		a.cfg.LeetCode.Site,
		a.cfg.LeetCode.Cookie,
		a.cfg.LeetCode.CSRFToken,
	)

	fmt.Println("Syncing problems from LeetCode...")

	skip := 0
	limit := 100
	total := -1

	for {
		items, count, err := lcClient.FetchProblemList(ctx, skip, limit)
		if err != nil {
			return fmt.Errorf("fetch problems at skip=%d: %w", skip, err)
		}

		if total < 0 {
			total = count
			fmt.Printf("Total problems: %d\n", total)
		}

		for _, item := range items {
			if err := queries.UpsertProblem(ctx, problemToUpsertParams(item, a.cfg.LeetCode.Site)); err != nil {
				fmt.Printf("Warning: failed to upsert problem %s: %v\n", item.TitleSlug, err)
			}
			for _, tag := range item.TopicTags {
				if err := queries.UpsertProblemTag(ctx, sqlc.UpsertProblemTagParams{
					ProblemID: sql.NullInt64{Int64: int64(item.ID), Valid: true},
					Tag:       tag,
				}); err != nil {
					fmt.Printf("Warning: failed to upsert tag %q for problem %s: %v\n", tag, item.TitleSlug, err)
				}
			}
		}

		skip += len(items)
		fmt.Printf("  Synced %d/%d\n", skip, total)

		if skip >= total || len(items) == 0 {
			break
		}
	}

	fmt.Printf("Sync complete. %d problems in database.\n", skip)
	return nil
}

func problemToUpsertParams(item leetcode.ProblemListItem, site string) sqlc.UpsertProblemParams {
	params := sqlc.UpsertProblemParams{
		ID:         int64(item.ID),
		Title:      item.Title,
		TitleSlug:  item.TitleSlug,
		FrontendID: item.FrontendID,
		Difficulty: item.Difficulty,
		IsPaidOnly: sql.NullInt64{Int64: boolToInt(item.IsPaidOnly), Valid: true},
		AcRate:     sql.NullFloat64{Float64: item.AcRate, Valid: true},
		Site:       sql.NullString{String: site, Valid: true},
	}
	if item.Status != "" {
		params.Status = sql.NullString{String: item.Status, Valid: true}
	}
	return params
}

func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}
