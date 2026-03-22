package tui

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"fmt"

	"github.com/tzy0608/leet-tui/internal/config"
	"github.com/tzy0608/leet-tui/internal/db/sqlc"
	"github.com/tzy0608/leet-tui/internal/leetcode"
	"github.com/tzy0608/leet-tui/internal/srs"
	"github.com/tzy0608/leet-tui/internal/study"
	"github.com/tzy0608/leet-tui/internal/tui/components/help"
	"github.com/tzy0608/leet-tui/internal/tui/components/statusbar"
	"github.com/tzy0608/leet-tui/internal/tui/components/topicpanel"
	"github.com/tzy0608/leet-tui/internal/tui/layout"
	"github.com/tzy0608/leet-tui/internal/tui/page"
	"github.com/tzy0608/leet-tui/internal/tui/theme"
)

// Model is the root TUI model that manages page routing and global state.
type Model struct {
	cfg         *config.Config
	db          *sql.DB
	queries     *sqlc.Queries
	lcClient    leetcode.Client
	fsrs        *srs.FSRS
	studyEngine *study.Engine

	// Pages
	dashboard     *page.Dashboard
	problems      *page.Problems
	problemDetail *page.ProblemDetail
	review        *page.Review
	plans         *page.Plans

	// UI
	statusBar   *statusbar.Bar
	helpOverlay *help.Overlay
	showHelp    bool
	activePage  page.ID

	// Window dimensions
	width  int
	height int
}

// New creates the root TUI model.
func New(cfg *config.Config, db *sql.DB, queries *sqlc.Queries, lcClient leetcode.Client) *Model {
	fsrsAlgo := srs.NewFSRS(cfg.SRS.Weights, cfg.SRS.RetentionRate)
	scheduler := srs.NewScheduler(fsrsAlgo, queries)
	engine := study.NewEngine(queries, scheduler, cfg.General.NewPerDay)

	m := &Model{
		cfg:           cfg,
		db:            db,
		queries:       queries,
		lcClient:      lcClient,
		fsrs:          fsrsAlgo,
		studyEngine:   engine,
		dashboard:     page.NewDashboard(),
		problems:      page.NewProblems(),
		problemDetail: page.NewProblemDetail(cfg.Editor.Command, cfg.General.Language),
		review:        page.NewReview(fsrsAlgo),
		plans:         page.NewPlans(),
		statusBar:     statusbar.New(),
		activePage:    page.DashboardPage,
	}
	m.helpOverlay = help.New(globalBindings())
	return m
}

func globalBindings() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "Dashboard")),
		key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "Problems")),
		key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "Review")),
		key.NewBinding(key.WithKeys("4"), key.WithHelp("4", "Plans")),
		key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "Toggle help")),
		key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q/ctrl+c", "Quit")),
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "Switch focus")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "Select")),
		key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "Back")),
		key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/↓", "Move down")),
		key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k/↑", "Move up")),
		key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "Search")),
		key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "Edit in $EDITOR")),
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.dashboard.Init(),
		m.seedPlans(),
		m.loadDashboard(),
	)
}

// seedPlans loads predefined plans into the database once at startup.
func (m *Model) seedPlans() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if err := m.studyEngine.LoadPredefinedPlans(ctx); err != nil {
			return page.ErrorMsg{Err: fmt.Errorf("seed plans: %w", err)}
		}
		return nil
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resize()

	case page.ErrorMsg:
		m.statusBar.SetStatus("Error: " + msg.Err.Error())
		return m, nil

	case page.ProblemsLoadedMsg:
		// Route to active page first, then trigger async status loading
		cmd := m.routeToActivePage(msg)
		cmds = append(cmds, cmd)
		cmds = append(cmds, m.loadProblemStatuses())
		return m, tea.Batch(cmds...)

	case page.ProblemDetailLoadedMsg:
		// Route to active page first, then trigger async accepted code loading
		cmd := m.routeToActivePage(msg)
		cmds = append(cmds, cmd)
		if msg.Problem != nil {
			cmds = append(cmds, m.loadLatestAcceptedCode(msg.Problem.TitleSlug))
		}
		return m, tea.Batch(cmds...)

	case page.RunTestRequestMsg:
		return m, m.runTest(msg)

	case page.SubmitRequestMsg:
		return m, m.submitCode(msg)

	case page.ChangeMsg:
		return m, m.navigateTo(msg)

	case page.PlanSelectedMsg:
		m.statusBar.SetStatus("Activating plan...")
		return m, m.activatePlan(msg.PlanID)

	case page.PlansLoadedMsg:
		m.statusBar.SetStatus("")
		cmd := m.routeToActivePage(msg)
		return m, cmd

	case tea.KeyMsg:
		inputFocused := m.isInputFocused()
		// Global key handling
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if !inputFocused {
				if m.activePage == page.DashboardPage || m.showHelp {
					return m, tea.Quit
				}
				// On non-Dashboard pages, q navigates back to Dashboard
				return m, m.navigateTo(page.ChangeMsg{Target: page.DashboardPage})
			}
		case "?":
			if !inputFocused {
				m.showHelp = !m.showHelp
				return m, nil
			}
		case "1":
			if !inputFocused {
				return m, m.navigateTo(page.ChangeMsg{Target: page.DashboardPage})
			}
		case "2":
			if !inputFocused {
				return m, m.navigateTo(page.ChangeMsg{Target: page.ProblemsPage})
			}
		case "3":
			if !inputFocused {
				return m, m.navigateTo(page.ChangeMsg{Target: page.ReviewPage})
			}
		case "4":
			if !inputFocused {
				return m, m.navigateTo(page.ChangeMsg{Target: page.PlansPage})
			}
		case "esc":
			if m.showHelp {
				m.showHelp = false
				return m, nil
			}
		}
	}

	// Route message to active page
	cmd := m.routeToActivePage(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) isInputFocused() bool {
	var activePage any
	switch m.activePage {
	case page.DashboardPage:
		activePage = m.dashboard
	case page.ProblemsPage:
		activePage = m.problems
	case page.ProblemDetailPage:
		activePage = m.problemDetail
	case page.ReviewPage:
		activePage = m.review
	case page.PlansPage:
		activePage = m.plans
	}
	if f, ok := activePage.(page.InputFocuser); ok {
		return f.IsInputFocused()
	}
	return false
}

func (m *Model) navigateTo(msg page.ChangeMsg) tea.Cmd {
	// Blur current page
	m.blurAll()

	m.activePage = msg.Target
	m.statusBar.SetPage(msg.Target.String())
	m.showHelp = false

	var cmds []tea.Cmd

	switch msg.Target {
	case page.DashboardPage:
		m.dashboard.Focus()
		cmds = append(cmds, m.loadDashboard())
	case page.ProblemsPage:
		m.problems.Focus()
		cmds = append(cmds, m.loadProblems())
	case page.ProblemDetailPage:
		m.problemDetail.Focus()
		if msg.Args != nil {
			if slug, ok := msg.Args["slug"].(string); ok {
				cmds = append(cmds, m.loadProblemDetail(slug))
			}
		}
	case page.ReviewPage:
		m.review.Focus()
		cmds = append(cmds, m.loadReviewQueue())
	case page.PlansPage:
		m.plans.Focus()
		cmds = append(cmds, m.loadPlans())
	}

	return tea.Batch(cmds...)
}

func (m *Model) blurAll() {
	m.dashboard.Blur()
	m.problems.Blur()
	m.problemDetail.Blur()
	m.review.Blur()
	m.plans.Blur()
}

func (m *Model) routeToActivePage(msg tea.Msg) tea.Cmd {
	var model tea.Model
	var cmd tea.Cmd

	switch m.activePage {
	case page.DashboardPage:
		model, cmd = m.dashboard.Update(msg)
		m.dashboard = model.(*page.Dashboard)
	case page.ProblemsPage:
		model, cmd = m.problems.Update(msg)
		m.problems = model.(*page.Problems)
	case page.ProblemDetailPage:
		model, cmd = m.problemDetail.Update(msg)
		m.problemDetail = model.(*page.ProblemDetail)
	case page.ReviewPage:
		model, cmd = m.review.Update(msg)
		m.review = model.(*page.Review)
	case page.PlansPage:
		model, cmd = m.plans.Update(msg)
		m.plans = model.(*page.Plans)
	}

	return cmd
}

func (m *Model) resize() {
	contentH := m.height - 1 // status bar

	m.dashboard.SetSize(m.width, contentH)
	m.problems.SetSize(m.width, contentH)
	m.problemDetail.SetSize(m.width, contentH)
	m.review.SetSize(m.width, contentH)
	m.plans.SetSize(m.width, contentH)
	m.statusBar.SetWidth(m.width)
	m.helpOverlay.SetSize(m.width, contentH)
}

func (m *Model) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	var pageView string
	switch m.activePage {
	case page.DashboardPage:
		pageView = m.dashboard.View()
	case page.ProblemsPage:
		pageView = m.problems.View()
	case page.ProblemDetailPage:
		pageView = m.problemDetail.View()
	case page.ReviewPage:
		pageView = m.review.View()
	case page.PlansPage:
		pageView = m.plans.View()
	}

	statusView := m.statusBar.View()

	contentH := m.height - 1
	pageArea := lipgloss.NewStyle().
		Width(m.width).
		Height(contentH).
		MaxHeight(contentH).
		Render(pageView)

	base := pageArea + "\n" + statusView

	// Overlay help
	if m.showHelp {
		helpView := m.helpOverlay.View()
		return layout.PlaceOverlay(m.width, m.height, helpView, base)
	}

	return base
}

// Data loading commands

func (m *Model) activatePlan(planID int) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if err := m.queries.SetActivePlan(ctx, int64(planID)); err != nil {
			return page.ErrorMsg{Err: fmt.Errorf("activate plan: %w", err)}
		}

		plans, progress, err := m.fetchPlansWithProgress(ctx)
		if err != nil {
			return page.ErrorMsg{Err: fmt.Errorf("reload plans: %w", err)}
		}
		return page.PlansLoadedMsg{Plans: plans, Progress: progress}
	}
}

func (m *Model) loadDashboard() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		now := time.Now()

		stats, err := study.GetStats(ctx, m.queries, m.cfg.GetSite())
		if err != nil {
			return page.ErrorMsg{Err: fmt.Errorf("load stats: %w", err)}
		}

		// Best-effort sync: ignore errors (user may not have cookie configured)
		_ = m.studyEngine.SyncPlanCompletion(ctx, m.lcClient)

		queue, err := m.studyEngine.BuildDailyQueue(ctx, now)
		if err != nil {
			return page.ErrorMsg{Err: fmt.Errorf("load queue: %w", err)}
		}

		return page.DashboardLoadedMsg{Queue: queue, Stats: stats}
	}
}

func (m *Model) loadProblems() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		problems, err := m.queries.ListProblems(ctx, sql.NullString{String: m.cfg.GetSite(), Valid: true})
		if err != nil {
			return page.ErrorMsg{Err: fmt.Errorf("load problems: %w", err)}
		}

		// Convert sqlc.Problem to leetcode.ProblemListItem
		var items []leetcode.ProblemListItem
		for _, p := range problems {
			item := leetcode.ProblemListItem{
				ID:         int(p.ID),
				Title:      p.Title,
				TitleSlug:  p.TitleSlug,
				FrontendID: p.FrontendID,
				Difficulty: p.Difficulty,
				IsPaidOnly: p.IsPaidOnly.Int64 != 0,
			}
			if p.AcRate.Valid {
				item.AcRate = float64(p.AcRate.Float64)
			}
			if p.Status.Valid {
				item.Status = p.Status.String
			}
			if p.TopicTags.Valid {
				var tags []string
				if err := json.Unmarshal([]byte(p.TopicTags.String), &tags); err == nil {
					item.TopicTags = tags
				}
			}
			items = append(items, item)
		}

		// Load tags
		tagRows, err := m.queries.ListTags(ctx)
		var tags []topicpanel.TagItem
		if err == nil {
			for _, t := range tagRows {
				tags = append(tags, topicpanel.TagItem{Name: t.Tag, Count: int(t.Cnt)})
			}
		}

		return page.ProblemsLoadedMsg{
			Items: items,
			Tags:  tags,
			Total: len(items),
		}
	}
}

func (m *Model) loadProblemStatuses() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		items, _, err := m.lcClient.FetchProblemList(ctx, 0, 100)
		if err != nil {
			return nil // silently ignore
		}

		statuses := make(map[string]string, len(items))
		for _, item := range items {
			if item.Status != "" {
				statuses[item.TitleSlug] = item.Status
			}
		}

		// Persist statuses to DB
		for slug, status := range statuses {
			_ = m.queries.UpdateProblemStatus(ctx, sqlc.UpdateProblemStatusParams{
				Status:    sql.NullString{String: status, Valid: true},
				TitleSlug: slug,
			})
		}

		if len(statuses) == 0 {
			return nil
		}
		return page.ProblemStatusUpdatedMsg{Statuses: statuses}
	}
}

func (m *Model) loadProblemDetail(slug string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Try DB first
		dbProb, err := m.queries.GetProblemBySlug(ctx, slug)
		if err == nil && dbProb.Content.Valid && dbProb.Content.String != "" {
			p := dbProblemToLeetcode(dbProb)
			return page.ProblemDetailLoadedMsg{Problem: p}
		}

		// Fetch from LeetCode API
		p, err := m.lcClient.FetchProblemDetail(ctx, slug)
		if err != nil {
			return page.ErrorMsg{Err: fmt.Errorf("fetch problem detail: %w", err)}
		}

		// Cache to DB
		tagsJSON, _ := json.Marshal(p.TopicTags)
		snippetsJSON, _ := json.Marshal(p.CodeSnippets)
		_ = m.queries.UpsertProblem(ctx, sqlc.UpsertProblemParams{
			ID:           int64(p.ID),
			Title:        p.Title,
			TitleSlug:    p.TitleSlug,
			FrontendID:   p.FrontendID,
			Difficulty:   p.Difficulty,
			Content:      sql.NullString{String: p.Content, Valid: true},
			TopicTags:    sql.NullString{String: string(tagsJSON), Valid: true},
			CodeSnippets: sql.NullString{String: string(snippetsJSON), Valid: true},
			IsPaidOnly:   sql.NullInt64{Int64: boolToInt64(p.IsPaidOnly), Valid: true},
			AcRate:       sql.NullFloat64{Float64: p.AcRate, Valid: true},
			Site:         sql.NullString{String: m.cfg.LeetCode.Site, Valid: true},
		})

		return page.ProblemDetailLoadedMsg{Problem: p}
	}
}

func (m *Model) loadLatestAcceptedCode(slug string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		langSlug := leetcode.LangSlug(m.cfg.General.Language)
		code, err := m.lcClient.FetchLatestAcceptedCode(ctx, slug, langSlug)
		if err != nil || code == "" {
			return nil
		}
		return page.AcceptedCodeLoadedMsg{Slug: slug, Code: code}
	}
}

func (m *Model) runTest(req page.RunTestRequestMsg) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		result, err := m.lcClient.RunCode(ctx, req.TitleSlug, req.QuestionID, req.Lang, req.Code, req.TestInput)
		return page.RunResultMsg{Result: result, Err: err}
	}
}

func (m *Model) submitCode(req page.SubmitRequestMsg) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		result, err := m.lcClient.SubmitCode(ctx, req.TitleSlug, req.QuestionID, req.Lang, req.Code)
		return page.SubmitResultMsg{Result: result, Err: err}
	}
}

func (m *Model) loadReviewQueue() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		now := time.Now()

		dueCards, err := m.queries.ListDueReviewCards(ctx, now)
		if err != nil {
			return page.ErrorMsg{Err: fmt.Errorf("load review queue: %w", err)}
		}

		var items []page.ReviewItem
		for _, dc := range dueCards {
			card := srs.CardFromDB(sqlc.ReviewCard{
				ProblemID:     dc.ProblemID,
				Due:           dc.Due,
				Stability:     dc.Stability,
				Difficulty:    dc.Difficulty,
				ElapsedDays:   dc.ElapsedDays,
				ScheduledDays: dc.ScheduledDays,
				Reps:          dc.Reps,
				Lapses:        dc.Lapses,
				State:         dc.State,
				LastReview:    dc.LastReview,
			})

			prob := &leetcode.Problem{
				ID:         int(dc.ProblemID),
				Title:      dc.Title,
				FrontendID: dc.FrontendID,
				Difficulty: dc.ProbDifficulty,
				TitleSlug:  dc.TitleSlug,
			}

			// Try to load full problem content for the review view
			if dbProb, err := m.queries.GetProblem(ctx, dc.ProblemID); err == nil {
				prob = dbProblemToLeetcode(dbProb)
			}

			items = append(items, page.ReviewItem{Card: card, Problem: prob})
		}

		return page.ReviewQueueMsg{Items: items}
	}
}

func (m *Model) loadPlans() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		plans, progress, err := m.fetchPlansWithProgress(ctx)
		if err != nil {
			return page.ErrorMsg{Err: fmt.Errorf("load plans: %w", err)}
		}
		return page.PlansLoadedMsg{Plans: plans, Progress: progress}
	}
}

func (m *Model) fetchPlansWithProgress(ctx context.Context) ([]study.Plan, map[int]study.PlanProgress, error) {
	dbPlans, err := m.queries.ListStudyPlans(ctx)
	if err != nil {
		return nil, nil, err
	}

	var plans []study.Plan
	progress := make(map[int]study.PlanProgress)

	for _, sp := range dbPlans {
		plan := study.Plan{
			ID:           int(sp.ID),
			Name:         sp.Name,
			Slug:         sp.Slug,
			IsPredefined: sp.IsPredefined.Int64 != 0,
			IsActive:     sp.IsActive.Int64 != 0,
		}
		if sp.Description.Valid {
			plan.Description = sp.Description.String
		}
		plans = append(plans, plan)

		prog, err := m.queries.CountPlanProgress(ctx, sql.NullInt64{Int64: sp.ID, Valid: true})
		if err == nil {
			progress[int(sp.ID)] = study.PlanProgress{
				Total:     int(prog.Total),
				Completed: int(prog.Completed.Float64),
			}
		}
	}

	return plans, progress, nil
}

// Helper to convert sqlc.Problem to leetcode.Problem
func dbProblemToLeetcode(p sqlc.Problem) *leetcode.Problem {
	prob := &leetcode.Problem{
		ID:         int(p.ID),
		Title:      p.Title,
		TitleSlug:  p.TitleSlug,
		FrontendID: p.FrontendID,
		Difficulty: p.Difficulty,
		IsPaidOnly: p.IsPaidOnly.Int64 != 0,
	}
	if p.Content.Valid {
		prob.Content = p.Content.String
	}
	if p.AcRate.Valid {
		prob.AcRate = p.AcRate.Float64
	}
	if p.TopicTags.Valid {
		var tags []string
		if err := json.Unmarshal([]byte(p.TopicTags.String), &tags); err == nil {
			prob.TopicTags = tags
		}
	}
	if p.CodeSnippets.Valid {
		prob.CodeSnippets = make(map[string]string)
		_ = json.Unmarshal([]byte(p.CodeSnippets.String), &prob.CodeSnippets)
	}
	return prob
}

func boolToInt64(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// SetStatusMsg updates the status bar message.
func (m *Model) SetStatusMsg(msg string) {
	m.statusBar.SetStatus(msg)
}

// SetTheme applies a color theme.
func (m *Model) SetTheme(name string) {
	switch name {
	case "default":
		theme.Current = theme.DefaultColors()
	}
}
