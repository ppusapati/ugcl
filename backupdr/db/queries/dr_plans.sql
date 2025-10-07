-- name: CreateDRPlan :one
INSERT INTO disaster_recovery_plans (
    name, description, plan_type, severity, scope, rto, rpo, mttr,
    availability, pre_recovery_steps, recovery_steps, post_recovery_steps,
    rollback_steps, resource_requirements, dependencies, prerequisites,
    last_tested, test_schedule, validation_criteria, notification_list,
    escalation_matrix, communication_plan, is_active, version,
    approved_by, approved_at, created_by, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
    $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28
) RETURNING *;

-- name: GetDRPlanByID :one
SELECT * FROM disaster_recovery_plans WHERE id = $1;

-- name: UpdateDRPlan :exec
UPDATE disaster_recovery_plans SET
    name = $2, description = $3, plan_type = $4, severity = $5, scope = $6,
    rto = $7, rpo = $8, mttr = $9, availability = $10, pre_recovery_steps = $11,
    recovery_steps = $12, post_recovery_steps = $13, rollback_steps = $14,
    resource_requirements = $15, dependencies = $16, prerequisites = $17,
    test_schedule = $18, validation_criteria = $19, notification_list = $20,
    escalation_matrix = $21, communication_plan = $22, is_active = $23,
    version = $24, approved_by = $25, approved_at = $26, updated_by = $27,
    metadata = $28
WHERE id = $1;

-- name: DeleteDRPlan :exec
DELETE FROM disaster_recovery_plans WHERE id = $1;

-- name: ListDRPlans :many
SELECT * FROM disaster_recovery_plans
WHERE ($1::text IS NULL OR plan_type = $1::text)
  AND ($2::text IS NULL OR severity = $2::text)
  AND ($3::text IS NULL OR scope = $3::text)
  AND ($4::boolean IS NULL OR is_active = $4::boolean)
  AND ($5::integer IS NULL OR version = $5::integer)
ORDER BY created_at DESC;

-- name: GetActiveDRPlans :many
SELECT * FROM disaster_recovery_plans
WHERE is_active = true
ORDER BY severity DESC, created_at DESC;

-- name: GetDRPlansByType :many
SELECT * FROM disaster_recovery_plans
WHERE plan_type = $1 AND is_active = true
ORDER BY severity DESC;

-- name: UpdateDRPlanLastTested :exec
UPDATE disaster_recovery_plans
SET last_tested = $2, updated_at = NOW()
WHERE id = $1;