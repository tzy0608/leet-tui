-- name: GetReviewCard :one
SELECT * FROM review_cards WHERE problem_id = ? LIMIT 1;

-- name: UpsertReviewCard :exec
INSERT INTO review_cards (problem_id, due, stability, difficulty, elapsed_days, scheduled_days, reps, lapses, state, last_review)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(problem_id) DO UPDATE SET
    due = excluded.due,
    stability = excluded.stability,
    difficulty = excluded.difficulty,
    elapsed_days = excluded.elapsed_days,
    scheduled_days = excluded.scheduled_days,
    reps = excluded.reps,
    lapses = excluded.lapses,
    state = excluded.state,
    last_review = excluded.last_review;

-- name: ListDueReviewCards :many
SELECT rc.*, p.title, p.frontend_id, p.difficulty as prob_difficulty, p.title_slug
FROM review_cards rc
JOIN problems p ON rc.problem_id = p.id
WHERE rc.due <= ?
ORDER BY rc.due ASC;

-- name: CountDueReviewCards :one
SELECT COUNT(*) FROM review_cards WHERE due <= ?;

-- name: CreateReviewLog :exec
INSERT INTO review_logs (problem_id, rating, state, elapsed_days, scheduled_days, time_spent_sec)
VALUES (?, ?, ?, ?, ?, ?);

-- name: ListReviewLogsByProblem :many
SELECT * FROM review_logs WHERE problem_id = ? ORDER BY reviewed_at DESC;

-- name: CountReviewsToday :one
SELECT COUNT(*) FROM review_logs WHERE DATE(reviewed_at) = DATE('now');
