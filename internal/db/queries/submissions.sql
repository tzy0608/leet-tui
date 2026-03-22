-- name: CreateSubmission :exec
INSERT INTO submissions (problem_id, lang, code, status, runtime_ms, memory_kb)
VALUES (?, ?, ?, ?, ?, ?);

-- name: ListSubmissionsByProblem :many
SELECT * FROM submissions WHERE problem_id = ? ORDER BY submitted_at DESC;

-- name: CountAcceptedProblems :one
SELECT COUNT(DISTINCT problem_id) FROM submissions WHERE status = 'Accepted';

-- name: CountAcceptedByDifficulty :many
SELECT p.difficulty, COUNT(DISTINCT s.problem_id) as cnt
FROM submissions s
JOIN problems p ON s.problem_id = p.id
WHERE s.status = 'Accepted'
GROUP BY p.difficulty;
