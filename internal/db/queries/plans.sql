-- name: CreateStudyPlan :one
INSERT INTO study_plans (name, slug, description, is_predefined, is_active)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetStudyPlan :one
SELECT * FROM study_plans WHERE id = ? LIMIT 1;

-- name: GetStudyPlanBySlug :one
SELECT * FROM study_plans WHERE slug = ? LIMIT 1;

-- name: ListStudyPlans :many
SELECT * FROM study_plans ORDER BY is_active DESC, created_at DESC;

-- name: SetActivePlan :exec
UPDATE study_plans SET is_active = CASE WHEN id = ? THEN 1 ELSE 0 END;

-- name: DeactivateAllPlans :exec
UPDATE study_plans SET is_active = 0;

-- name: AddProblemToPlan :exec
INSERT INTO study_plan_problems (plan_id, problem_id, day_number, topic_group, sort_order)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT DO NOTHING;

-- name: ListPlanProblems :many
SELECT spp.*, p.title, p.frontend_id, p.difficulty, p.title_slug
FROM study_plan_problems spp
JOIN problems p ON spp.problem_id = p.id
WHERE spp.plan_id = ?
ORDER BY spp.day_number, spp.sort_order;

-- name: CompletePlanProblem :exec
UPDATE study_plan_problems
SET is_completed = 1, completed_at = CURRENT_TIMESTAMP
WHERE plan_id = ? AND problem_id = ?;

-- name: CountPlanProgress :one
SELECT
    COUNT(*) as total,
    SUM(CASE WHEN is_completed = 1 THEN 1 ELSE 0 END) as completed
FROM study_plan_problems WHERE plan_id = ?;

-- name: GetActivePlan :one
SELECT * FROM study_plans WHERE is_active = 1 LIMIT 1;

-- name: ListIncompletePlanProblems :many
SELECT spp.*, p.title, p.frontend_id, p.difficulty, p.title_slug
FROM study_plan_problems spp
JOIN problems p ON spp.problem_id = p.id
WHERE spp.plan_id = ? AND spp.is_completed = 0
ORDER BY spp.day_number, spp.sort_order
LIMIT ?;
