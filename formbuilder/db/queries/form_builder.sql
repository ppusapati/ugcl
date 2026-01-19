-- =============================================================================
-- queries.sql - All SQLC queries for form builder
-- =============================================================================

-- =============================================================================
-- FORM QUERIES
-- =============================================================================

-- name: CreateForm :one
INSERT INTO forms (
    id, title, description, version, created_by, allowed_roles, audit,
    created_at, updated_at, table_name, module, schema_version, core_fields,
    steps, dependencies, cross_field_validations, workflow_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
) RETURNING *;

-- name: GetForm :one
SELECT * FROM forms 
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetFormByVersion :one
SELECT * FROM forms 
WHERE id = $1 AND version = $2 AND deleted_at IS NULL;

-- name: UpdateForm :one
UPDATE forms SET
    title = $2,
    description = $3,
    version = $4,
    allowed_roles = $5,
    audit = $6,
    updated_at = $7,
    table_name = $8,
    schema_version = $9,
    core_fields = $10,
    steps = $11,
    dependencies = $12,
    cross_field_validations = $13,
    workflow_id = $14
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteForm :exec
UPDATE forms SET deleted_at = NOW() WHERE id = $1;

-- name: ListForms :many
SELECT * FROM forms 
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountForms :one
SELECT COUNT(*) FROM forms WHERE deleted_at IS NULL;

-- name: SearchForms :many
SELECT * FROM forms 
WHERE deleted_at IS NULL 
  AND (title ILIKE '%' || $1 || '%' OR description ILIKE '%' || $1 || '%')
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetFormsByCreator :many
SELECT * FROM forms 
WHERE created_by = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;


-- =============================================================================
-- COMPLEX QUERIES FOR ANALYTICS AND REPORTING
-- =============================================================================

-- name: GetFormStatistics :one
SELECT 
    f.id,
    f.title,
    COUNT(fi.id) as total_instances,
    COUNT(CASE WHEN fi.current_state = 'completed' THEN 1 END) as completed_instances,
    COUNT(CASE WHEN fi.current_state = 'draft' THEN 1 END) as draft_instances,
    COUNT(CASE WHEN fi.current_state = 'in_review' THEN 1 END) as in_review_instances,
    AVG(EXTRACT(EPOCH FROM (fi.updated_at - fi.created_at))) as avg_completion_time_seconds
FROM forms f
LEFT JOIN form_instances fi ON f.id = fi.form_id AND fi.deleted_at IS NULL
WHERE f.id = $1 AND f.deleted_at IS NULL
GROUP BY f.id, f.title;

-- name: GetUserWorkload :many
SELECT 
    assigned_to,
    COUNT(*) as assigned_count,
    COUNT(CASE WHEN current_state = 'pending_approval' THEN 1 END) as pending_approval_count,
    COUNT(CASE WHEN current_state = 'in_review' THEN 1 END) as in_review_count
FROM form_instances 
WHERE assigned_to IS NOT NULL AND deleted_at IS NULL
GROUP BY assigned_to
ORDER BY assigned_count DESC;

-- name: GetOverdueInstances :many
SELECT fi.*, f.title as form_title
FROM form_instances fi
JOIN forms f ON fi.form_id = f.id
WHERE fi.deleted_at IS NULL 
  AND f.deleted_at IS NULL
  AND fi.created_at < NOW() - INTERVAL '7 days'
  AND fi.current_state NOT IN ('completed', 'cancelled')
ORDER BY fi.created_at ASC;

-- name: GetInstancesRequiringEscalation :many
SELECT fi.*, f.title as form_title
FROM form_instances fi
JOIN forms f ON fi.form_id = f.id
WHERE fi.deleted_at IS NULL 
  AND f.deleted_at IS NULL
  AND fi.updated_at < NOW() - INTERVAL '3 days'
  AND fi.current_state IN ('pending_approval', 'in_review')
ORDER BY fi.updated_at ASC;

-- name: GetRecentActivity :many
SELECT 
    al.id,
    al.action,
    al.timestamp,
    al.user_id,
    fi.id as instance_id,
    f.title as form_title
FROM audit_logs al
JOIN form_instances fi ON al.instance_id = fi.id
JOIN forms f ON fi.form_id = f.id
WHERE al.timestamp >= $1
  AND fi.deleted_at IS NULL
  AND f.deleted_at IS NULL
ORDER BY al.timestamp DESC
LIMIT $2 OFFSET $3;