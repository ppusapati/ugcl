
-- =============================================================================
-- SLA RULE QUERIES
-- =============================================================================

-- name: CreateSLARule :one
INSERT INTO sla_rules (
    id, name, state, duration, escalation_levels, active,
    applicable_roles, condition, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: GetSLARule :one
SELECT * FROM sla_rules 
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetSLARules :many
SELECT * FROM sla_rules 
WHERE deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetSLARulesByState :many
SELECT * FROM sla_rules 
WHERE state = $1 AND active = true AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetActiveSLARules :many
SELECT * FROM sla_rules 
WHERE active = true AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: UpdateSLARule :one
UPDATE sla_rules SET
    name = $2,
    state = $3,
    duration = $4,
    escalation_levels = $5,
    active = $6,
    applicable_roles = $7,
    condition = $8,
    updated_at = $9
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteSLARule :exec
UPDATE sla_rules SET deleted_at = NOW() WHERE id = $1;

-- name: ToggleSLARule :one
UPDATE sla_rules SET
    active = NOT active,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;
