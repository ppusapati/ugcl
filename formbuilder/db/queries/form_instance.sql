-- =============================================================================
-- FORM INSTANCE QUERIES
-- =============================================================================

-- name: CreateFormInstance :one
INSERT INTO form_instances (
    id, form_id, current_state, field_values, created_at, updated_at,
    created_by, assigned_to, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetFormInstance :one
SELECT * FROM form_instances 
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateFormInstance :one
UPDATE form_instances SET
    current_state = $2,
    field_values = $3,
    updated_at = $4,
    assigned_to = $5,
    metadata = $6
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteFormInstance :exec
UPDATE form_instances SET deleted_at = NOW() WHERE id = $1;

-- name: ListFormInstances :many
SELECT * FROM form_instances 
WHERE form_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountFormInstances :one
SELECT COUNT(*) FROM form_instances 
WHERE form_id = $1 AND deleted_at IS NULL;

-- name: GetFormInstancesByState :many
SELECT * FROM form_instances 
WHERE form_id = $1 AND current_state = $2 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: GetFormInstancesByAssignee :many
SELECT * FROM form_instances 
WHERE assigned_to = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetFormInstancesByCreator :many
SELECT * FROM form_instances 
WHERE created_by = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateFormInstanceState :one
UPDATE form_instances SET
    current_state = $2,
    updated_at = NOW(),
    assigned_to = $3
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;
