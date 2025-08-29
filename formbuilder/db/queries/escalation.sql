

-- =============================================================================
-- ESCALATION QUERIES
-- =============================================================================

-- name: CreateEscalation :one
INSERT INTO escalations (
    id, name, trigger_condition, from_states, to_state, notify_roles,
    escalation_message, auto_escalate, after_duration, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: GetEscalation :one
SELECT * FROM escalations 
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetEscalations :many
SELECT * FROM escalations 
WHERE deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetEscalationsByFromState :many
SELECT * FROM escalations 
WHERE from_states @> $1::jsonb AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetAutoEscalations :many
SELECT * FROM escalations 
WHERE auto_escalate = true AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: UpdateEscalation :one
UPDATE escalations SET
    name = $2,
    trigger_condition = $3,
    from_states = $4,
    to_state = $5,
    notify_roles = $6,
    escalation_message = $7,
    auto_escalate = $8,
    after_duration = $9,
    updated_at = $10
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteEscalation :exec
UPDATE escalations SET deleted_at = NOW() WHERE id = $1;

-- name: ToggleAutoEscalation :one
UPDATE escalations SET
    auto_escalate = NOT auto_escalate,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;
