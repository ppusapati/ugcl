-- name: CreateRestoreRequest :one
INSERT INTO restore_requests (
    backup_job_id, requested_by, restore_type, restore_scope,
    target_location, overwrite_policy, include_filters, exclude_filters,
    point_in_time, status, progress_percent, restored_files,
    restored_size, error_count, error_message, requires_approval,
    approval_status, approved_by, approved_at, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
    $16, $17, $18, $19, $20
) RETURNING *;

-- name: GetRestoreRequestByID :one
SELECT * FROM restore_requests WHERE id = $1;

-- name: UpdateRestoreRequest :exec
UPDATE restore_requests SET
    restore_type = $2, restore_scope = $3, target_location = $4,
    overwrite_policy = $5, include_filters = $6, exclude_filters = $7,
    point_in_time = $8, status = $9, started_at = $10, completed_at = $11,
    progress_percent = $12, restored_files = $13, restored_size = $14,
    error_count = $15, error_message = $16, approval_status = $17,
    approved_by = $18, approved_at = $19, metadata = $20
WHERE id = $1;

-- name: ListRestoreRequests :many
SELECT * FROM restore_requests
WHERE ($1::uuid IS NULL OR backup_job_id = $1::uuid)
  AND ($2::uuid IS NULL OR requested_by = $2::uuid)
  AND ($3::text IS NULL OR status = $3::text)
  AND ($4::text IS NULL OR restore_type = $4::text)
  AND ($5::text IS NULL OR restore_scope = $5::text)
  AND ($6::boolean IS NULL OR requires_approval = $6::boolean)
  AND ($7::timestamptz IS NULL OR created_at >= $7::timestamptz)
  AND ($8::timestamptz IS NULL OR created_at <= $8::timestamptz)
ORDER BY created_at DESC;

-- name: GetPendingRestoreRequests :many
SELECT * FROM restore_requests
WHERE status = 'REQUESTED'
ORDER BY created_at ASC;

-- name: GetRestoreRequestsByUser :many
SELECT * FROM restore_requests
WHERE requested_by = $1
ORDER BY created_at DESC;

-- name: UpdateRestoreStatus :exec
UPDATE restore_requests
SET status = $2, started_at = CASE WHEN $2 = 'RUNNING' THEN NOW() ELSE started_at END,
    completed_at = CASE WHEN $2 IN ('COMPLETED', 'FAILED', 'CANCELLED') THEN NOW() ELSE completed_at END
WHERE id = $1;