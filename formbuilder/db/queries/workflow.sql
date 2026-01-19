

-- =============================================================================
-- WORKFLOW QUERIES
-- =============================================================================

-- name: CreateWorkflow :one
INSERT INTO workflows (
    id, initial_state, states, global_transitions, metadata,
    created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetWorkflow :one
SELECT * FROM workflows 
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetWorkflowByFormID :one
SELECT w.* FROM workflows w
JOIN forms f ON f.workflow_id = w.id
WHERE f.id = $1 AND w.deleted_at IS NULL AND f.deleted_at IS NULL;

-- name: UpdateWorkflow :one
UPDATE workflows SET
    initial_state = $2,
    states = $3,
    global_transitions = $4,
    metadata = $5,
    updated_at = $6
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteWorkflow :exec
UPDATE workflows SET deleted_at = NOW() WHERE id = $1;

-- name: ListWorkflows :many
SELECT * FROM workflows 
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountWorkflows :one
SELECT COUNT(*) FROM workflows WHERE deleted_at IS NULL;