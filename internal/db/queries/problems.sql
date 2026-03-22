-- name: GetProblem :one
SELECT * FROM problems WHERE id = ? LIMIT 1;

-- name: GetProblemBySlug :one
SELECT * FROM problems WHERE title_slug = ? LIMIT 1;

-- name: ListProblems :many
SELECT * FROM problems WHERE site = ? ORDER BY CAST(frontend_id AS INTEGER);

-- name: ListProblemsByDifficulty :many
SELECT * FROM problems WHERE site = ? AND difficulty = ? ORDER BY CAST(frontend_id AS INTEGER);

-- name: SearchProblems :many
SELECT * FROM problems
WHERE site = ? AND (title LIKE '%' || ? || '%' OR frontend_id = ?)
ORDER BY CAST(frontend_id AS INTEGER)
LIMIT 50;

-- name: UpsertProblem :exec
INSERT INTO problems (id, title, title_slug, frontend_id, difficulty, content, topic_tags, code_snippets, is_paid_only, ac_rate, site, status, fetched_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(title_slug) DO UPDATE SET
    title = excluded.title,
    frontend_id = excluded.frontend_id,
    difficulty = excluded.difficulty,
    content = excluded.content,
    topic_tags = excluded.topic_tags,
    code_snippets = excluded.code_snippets,
    is_paid_only = excluded.is_paid_only,
    ac_rate = excluded.ac_rate,
    status = COALESCE(excluded.status, problems.status),
    fetched_at = CURRENT_TIMESTAMP;

-- name: CountProblems :one
SELECT COUNT(*) FROM problems WHERE site = ?;

-- name: CountProblemsByDifficulty :one
SELECT COUNT(*) FROM problems WHERE site = ? AND difficulty = ?;

-- name: UpsertProblemTag :exec
INSERT INTO problem_tags (problem_id, tag) VALUES (?, ?)
ON CONFLICT DO NOTHING;

-- name: ListTags :many
SELECT tag, COUNT(*) as cnt FROM problem_tags GROUP BY tag ORDER BY cnt DESC;

-- name: ListProblemsByTag :many
SELECT p.* FROM problems p
JOIN problem_tags pt ON p.id = pt.problem_id
WHERE pt.tag = ? AND p.site = ?
ORDER BY CAST(p.frontend_id AS INTEGER);

-- name: CountProblemsByDifficultyAll :many
SELECT difficulty, COUNT(*) as cnt FROM problems WHERE site = ? GROUP BY difficulty;

-- name: CountSolvedProblems :one
SELECT COUNT(*) FROM problems WHERE status = 'ac' AND site = ?;

-- name: CountSolvedByDifficulty :many
SELECT difficulty, COUNT(*) as cnt FROM problems WHERE status = 'ac' AND site = ? GROUP BY difficulty;

-- name: UpdateProblemStatus :exec
UPDATE problems SET status = ? WHERE title_slug = ?;
