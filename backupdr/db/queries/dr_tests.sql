-- name: CreateDRTest :one
INSERT INTO recovery_executions (
    plan_id, execution_type, trigger_reason, status, started_at,
    completed_at, estimated_rto, actual_rto, total_steps, completed_steps,
    failed_steps, current_step, success_rate, data_recovered,
    systems_recovered, notifications_sent, stakeholders_notified,
    issues_encountered, resolution_actions, executed_by, execution_log, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
    $16, $17, $18, $19, $20, $21, $22
) RETURNING *;

-- name: GetDRTestByID :one
SELECT * FROM recovery_executions WHERE id = $1;

-- name: GetTestsByPlan :many
SELECT * FROM recovery_executions
WHERE plan_id = $1
ORDER BY started_at DESC;

-- name: UpdateDRTest :exec
UPDATE recovery_executions SET
    status = $2, completed_at = $3, actual_rto = $4, total_steps = $5,
    completed_steps = $6, failed_steps = $7, current_step = $8,
    success_rate = $9, data_recovered = $10, systems_recovered = $11,
    notifications_sent = $12, stakeholders_notified = $13,
    issues_encountered = $14, resolution_actions = $15, execution_log = $16,
    metadata = $17
WHERE id = $1;

-- name: ListDRTests :many
SELECT * FROM recovery_executions
WHERE ($1::uuid IS NULL OR plan_id = $1::uuid)
  AND ($2::text IS NULL OR execution_type = $2::text)
  AND ($3::text IS NULL OR status = $3::text)
  AND ($4::timestamptz IS NULL OR started_at >= $4::timestamptz)
  AND ($5::timestamptz IS NULL OR started_at <= $5::timestamptz)
  AND ($6::uuid IS NULL OR executed_by = $6::uuid)
ORDER BY started_at DESC;

-- name: GetOverdueTests :many
SELECT re.* FROM recovery_executions re
JOIN disaster_recovery_plans drp ON re.plan_id = drp.id
WHERE re.status = 'INITIATED'
  AND re.started_at <= $1
  AND drp.is_active = true
ORDER BY re.started_at ASC;

-- name: GetTestsByType :many
SELECT * FROM recovery_executions
WHERE execution_type = $1
ORDER BY started_at DESC
LIMIT $2;